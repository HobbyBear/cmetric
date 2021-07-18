package main

import (
	"cmetric"
	"fmt"
	"time"
)

func main() {

	for {
		cup := cmetric.CurrentCpuUsage()
		fmt.Println("cpu is ",cup)
		memory := cmetric.CurrentMemoryUsage()
		fmt.Println("memory is ",memory/(1024*1024) ," M")
		time.Sleep(2 * time.Second)
	}

}
