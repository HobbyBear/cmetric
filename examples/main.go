package main

import (
	"cmetric"
	"fmt"
	"time"
)

func main() {

	for {
		cup := cmetric.CurrentCpuUsage()
		fmt.Println("cpu ", cup)
		time.Sleep(2000 * time.Millisecond)
	}

}
