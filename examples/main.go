package main

import (
	"cmetric"
	"fmt"
	"time"
)

func main() {

	for {
		cup := cmetric.CurrentCpuUsage()
		if cup >= 0.5 {
			fmt.Println("cpu ", cup)
		}
		time.Sleep(100 * time.Millisecond)
	}

}
