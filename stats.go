package cmetric

import (
	"bufio"
	"errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	NotRetrievedLoadValue     float64 = -1.0
	NotRetrievedCpuUsageValue float64 = -1.0
	NotRetrievedMemoryValue   int64   = -1
	CGroupPath                        = "/proc/1/cgroup"
	DockerPath                        = "/docker"
	KubepodsPath                      = "/kubepods"
)

var (
	currentLoad        atomic.Value
	currentCpuUsage    atomic.Value
	currentMemoryUsage atomic.Value

	memoryStatCollectorOnce sync.Once
	cpuStatCollectorOnce    sync.Once

	CurrentPID         = os.Getpid()
	currentProcess     atomic.Value
	currentProcessOnce sync.Once

	ssStopChan = make(chan struct{})

	isContainer             bool
	preSysTotalCpuUsage     atomic.Value
	preContainerCpuUsage    atomic.Value
	onlineContainerCpuCount float64
)

func init() {
	currentLoad.Store(NotRetrievedLoadValue)
	currentCpuUsage.Store(NotRetrievedCpuUsageValue)
	currentMemoryUsage.Store(NotRetrievedMemoryValue)

	log.Println("current pid ", CurrentPID)
	p, err := process.NewProcess(int32(CurrentPID))
	if err != nil {
		log.Fatal(err, "Fail to new process when initializing system metric", "pid", CurrentPID)
		return
	}
	currentProcessOnce.Do(func() {
		currentProcess.Store(p)
	})

	isContainer = isContainerRunning()
	if isContainer {
		log.Println("environment is  container")
		var (
			currentSysCpuTotal       float64
			currentContainerCpuTotal float64
		)

		currentSysCpuTotal, err = getSysCpuUsage()
		if err != nil {
			log.Fatal(err, "Fail to getSysCpuUsage when initializing system metric")
			return
		}
		currentContainerCpuTotal, err = getContainerCpuUsage()
		if err != nil {
			log.Fatal(err, "Fail to getContainerCpuUsage when initializing system metric")
			return
		}
		preContainerCpuUsage.Store(currentContainerCpuTotal)
		preSysTotalCpuUsage.Store(currentSysCpuTotal)
		onlineContainerCpuCount, err = getContainerCpuCount()
		if err != nil {
			log.Fatal(err, "Fail to getContainerCpuCount when initializing system metric")
			return
		}
	}
	go InitCpuCollector(15)
	go InitMemoryCollector(15)
}

func isContainerRunning() bool {
	f, err := os.Open(CGroupPath)
	if err != nil {
		return false
	}
	defer f.Close()
	buff := bufio.NewReader(f)
	for {
		line, _, err := buff.ReadLine()
		if err != nil {
			return false
		}
		if strings.Contains(string(line), DockerPath) ||
			strings.Contains(string(line), KubepodsPath) {
			return true
		}
	}
}

func getContainerCpuCount() (float64, error) {
	path := "/sys/fs/cgroup/cpuacct/cpuacct.usage_percpu"
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	usage, _, err := reader.ReadLine()
	if err != nil {
		return 0, err
	}
	perCpuUsages := strings.Fields(string(usage))

	return float64(len(perCpuUsages)), nil
}

func getSysCpuUsage() (float64, error) {
	var (
		currentSysCpuTotal float64
	)
	currentCpuStatArr, err := cpu.Times(false)
	if err != nil {
		return 0, err
	}
	for _, stat := range currentCpuStatArr {
		currentSysCpuTotal = stat.User + stat.System + stat.Idle + stat.Nice + stat.Iowait + stat.Irq +
			stat.Softirq + stat.Steal + stat.Guest + stat.GuestNice
	}
	return currentSysCpuTotal, nil
}

func getContainerCpuUsage() (float64, error) {
	path := "/sys/fs/cgroup/cpuacct/cpuacct.usage"
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	usage, _, err := reader.ReadLine()
	if err != nil {
		return 0, err
	}
	ns, err := strconv.ParseFloat(strings.TrimSpace(string(usage)), 64)
	if err != nil {
		return 0, err
	}
	return ns / 1e9, nil
}

func InitCpuCollector(intervalMs uint32) {
	if intervalMs == 0 {
		return
	}
	cpuStatCollectorOnce.Do(func() {

		retrieveAndUpdateCpuStat()

		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				retrieveAndUpdateCpuStat()
			case <-ssStopChan:
				ticker.Stop()
				return
			}
		}
	})
}

func retrieveAndUpdateCpuStat() {
	var (
		cpuPercent float64
		err        error
	)
	if isContainer {
		cpuPercent, err = GetContainerCpuStat()
		if err != nil {
			log.Println(err, "Fail to retrieve and update cpu statistic")
			return
		}
	} else {
		cpuPercent, err = getProcessCpuStat()
		if err != nil {
			log.Println(err, "Fail to retrieve and update cpu statistic")
			return
		}
	}

	currentCpuUsage.Store(cpuPercent)
}

func GetContainerCpuStat() (float64, error) {

	var (
		currentSysCpuTotal       float64
		currentContainerCpuTotal float64
		err                      error
	)

	currentSysCpuTotal, err = getSysCpuUsage()
	if err != nil {
		return 0, err
	}
	currentContainerCpuTotal, err = getContainerCpuUsage()
	if err != nil {
		return 0, err
	}

	preSysTotalCpu, ok := preSysTotalCpuUsage.Load().(float64)
	if !ok {
		return 0, errors.New("preSysTotalCpuUsage load is not float64")
	}
	preContainerCpu, ok := preContainerCpuUsage.Load().(float64)

	if !ok {
		return 0, errors.New("preContainerCpuUsage load is not float64")
	}

	preSysTotalCpuUsage.Store(currentSysCpuTotal)
	preContainerCpuUsage.Store(currentContainerCpuTotal)

	if currentSysCpuTotal-preSysTotalCpu == 0 {
		return 0, err
	}
	return (currentContainerCpuTotal - preContainerCpu) * 0.5 / (currentSysCpuTotal - preSysTotalCpu), err
}

func getProcessCpuStat() (float64, error) {
	curProcess := currentProcess.Load()
	if curProcess == nil {
		p, err := process.NewProcess(int32(CurrentPID))
		if err != nil {
			return 0, err
		}
		currentProcessOnce.Do(func() {
			currentProcess.Store(p)
		})
		curProcess = currentProcess.Load()
	}
	p := curProcess.(*process.Process)
	return p.Percent(0)
}

func CurrentCpuUsage() float64 {
	r, ok := currentCpuUsage.Load().(float64)
	if !ok {
		return NotRetrievedCpuUsageValue
	}
	return r
}

func CurrentMemoryUsage() int64 {
	bytes, ok := currentMemoryUsage.Load().(int64)
	if !ok {
		return NotRetrievedMemoryValue
	}

	return bytes
}

func InitMemoryCollector(intervalMs uint32) {
	if intervalMs == 0 {
		return
	}
	memoryStatCollectorOnce.Do(func() {
		// Initial memory retrieval.
		retrieveAndUpdateMemoryStat()

		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				retrieveAndUpdateMemoryStat()
			case <-ssStopChan:
				ticker.Stop()
				return
			}
		}
	})
}

func retrieveAndUpdateMemoryStat() {
	var (
		memoryUsedBytes int64
		err             error
	)
	if isContainer {
		memoryUsedBytes, err = GetContainerMemoryStat()
		if err != nil {
			log.Println(err, "Fail to retrieve and update container memory statistic")
			return
		}
	} else {
		memoryUsedBytes, err = GetProcessMemoryStat()
		if err != nil {
			log.Println(err, "Fail to retrieve and update memory statistic")
			return
		}
	}
	currentMemoryUsage.Store(memoryUsedBytes)
}

func GetProcessMemoryStat() (int64, error) {
	curProcess := currentProcess.Load()
	if curProcess == nil {
		p, err := process.NewProcess(int32(CurrentPID))
		if err != nil {
			return 0, err
		}
		currentProcessOnce.Do(func() {
			currentProcess.Store(p)
		})
		curProcess = currentProcess.Load()
	}
	p := curProcess.(*process.Process)
	memInfo, err := p.MemoryInfo()
	var rss int64
	if memInfo != nil {
		rss = int64(memInfo.RSS)
	}

	return rss, err
}

func GetContainerMemoryStat() (int64, error) {
	path := "/sys/fs/cgroup/memory/memory.usage_in_bytes"
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	usage, _, err := reader.ReadLine()
	if err != nil {
		return 0, err
	}
	ns, err := strconv.ParseInt(strings.TrimSpace(string(usage)), 10, 64)
	if err != nil {
		return 0, err
	}
	return ns, nil
}
