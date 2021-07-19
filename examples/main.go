package main

import (
	"fmt"
	"github.com/HobbyBear/cmetric"
	"time"
)

func cpuTest() {
	go func() {
		for {
			select {
			case <-time.After(6 * time.Second):
				fmt.Println("cpu down===========")
				goto sleep
			default:

			}
		}
	sleep:
		time.Sleep(6 * time.Second)
		fmt.Println("cpu up===========")
		cpuTest()
	}()
}

func main() {

	cpuTest()

	for {
		cpu := cmetric.CurrentCpuUsage()
		fmt.Println("cpu ", cpu)
		memory := cmetric.CurrentMemoryUsage()
		fmt.Println("memory ", memory/(1024*1024))
		time.Sleep(2000 * time.Millisecond)
	}

}
