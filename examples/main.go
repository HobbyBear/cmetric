package main

import (
	"cmetric"
	"fmt"
	"time"
)

func main() {

	go func() {
		t := time.NewTicker(6 * time.Second)
		for {
			select {
			case <-t.C:
				fmt.Println("cpu down===========")
				time.Sleep(6 * time.Second)
				t.Reset(6 * time.Second)
				fmt.Println("cpu up===========")
			default:

			}
		}
	}()

	for {
		cpu := cmetric.CurrentCpuUsage()
		fmt.Println("cpu ", cpu)
		memory := cmetric.CurrentMemoryUsage()
		fmt.Println("memory ", memory/(1024*1024))
		time.Sleep(2000 * time.Millisecond)
	}

}
