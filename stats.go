package cmetric

import (
	"bufio"
	"github.com/shirou/gopsutil/process"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	NotRetrievedCpuUsageValue float64 = -1.0
	NotRetrievedMemoryValue   float32 = -1
	CGroupPath                        = "/proc/1/cgroup"
	DockerPath                        = "/docker"
	KubepodsPath                      = "/kubepods"

	cgroupCpuQuotaPath  = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	cgroupCpuPeriodPath = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	memoryPath          = "/sys/fs/cgroup/memory/memory.usage_in_bytes"
)

var (
	cpuPercentUsage    atomic.Value
	memoryPercentUsage atomic.Value

	memoryStatCollectorOnce sync.Once
	cpuStatCollectorOnce    sync.Once

	CurrentPID         = os.Getpid()
	currentProcess     atomic.Value
	currentProcessOnce sync.Once

	ssStopChan = make(chan struct{})

	isContainer bool
	cpuCount    float64
)

func init() {
	cpuPercentUsage.Store(NotRetrievedCpuUsageValue)
	memoryPercentUsage.Store(NotRetrievedMemoryValue)

	p, err := process.NewProcess(int32(CurrentPID))
	if err != nil {
		log.Fatal(err, "Fail to new process when initializing system metric", "pid", CurrentPID)
		return
	}
	currentProcessOnce.Do(func() {
		currentProcess.Store(p)
	})

	isContainer = isContainerRunning()
	cpuCount = float64(runtime.NumCPU())
	if isContainer {
		log.Println("environment is  container")
		cpuCount, err = getContainerCpuCount()
		if err != nil {
			log.Fatal(err, "Fail to getContainerCpuCount when initializing system metric")
			return
		}
	}
	go InitCpuCollector(1000)
	go InitMemoryCollector(150)
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

	cpuPeriod, err := readUint(cgroupCpuPeriodPath)
	if err != nil {
		return 0, err
	}

	cpuQuota, err := readUint(cgroupCpuQuotaPath)
	if err != nil {
		return 0, err
	}
	cpuCore := float64(cpuQuota) / float64(cpuPeriod)

	return cpuCore, nil
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

	cpuPercent, err = getProcessCpuStat()
	if err != nil {
		log.Println(err)
		return
	}
	cpuPercentUsage.Store(cpuPercent / cpuCount)
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

func CurrentCpuPercentUsage() float64 {
	r, ok := cpuPercentUsage.Load().(float64)
	if !ok {
		return NotRetrievedCpuUsageValue
	}
	return r
}

func CurrentMemoryPercentUsage() float32 {
	r, ok := memoryPercentUsage.Load().(float32)
	if !ok {
		return NotRetrievedMemoryValue
	}
	return r
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
		err           error
		memoryLimit   int64
		memoryPercent float32
	)
	curProcess := currentProcess.Load()
	p := curProcess.(*process.Process)

	if isContainer {
		memoryLimit, err = GetContainerMemoryLimit()
		if err != nil {
			log.Println(err, "Fail to retrieve and update container memory statistic")
			return
		}
		memoryInfo, err := p.MemoryInfo()
		if err != nil {
			log.Println(err, "Fail to retrieve and update container memory statistic")
			return
		}
		memoryPercent = float32(memoryInfo.RSS) / float32(memoryLimit)
	} else {
		memoryPercent, err = p.MemoryPercent()
		if err != nil {
			log.Println(err, "Fail to retrieve and update container memory statistic")
			return
		}
	}
	memoryPercentUsage.Store(memoryPercent)
}

func GetContainerMemoryLimit() (int64, error) {
	usage, err := readUint(memoryPath)
	if err != nil {
		return 0, err
	}
	return int64(usage), nil
}
