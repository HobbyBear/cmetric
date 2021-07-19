package main

import (
	"fmt"
	"github.com/HobbyBear/cmetric"
	"time"
)

func cpuTest() {
	t := time.AfterFunc(6*time.Second, func() {
		time.Sleep(6 * time.Second)
		fmt.Println("cpu up===========")
		go cpuTest()
	})
	for {
		select {
		case <-t.C:
			fmt.Println("cpu down===========")
			return
		default:

		}
	}
}

func main() {

	go cpuTest()

	for {
		cpu := cmetric.CurrentCpuUsage()
		fmt.Println("cpu ", cpu)
		memory := cmetric.CurrentMemoryUsage()
		fmt.Println("memory ", memory/(1024*1024))
		time.Sleep(2000 * time.Millisecond)
	}

}
