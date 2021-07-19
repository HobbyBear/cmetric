## read metrics in container


if environment is container, the cpu ,memory is relative to container,
else  the metrics is relative to host.

### usage

```

cpu := cmetric.CurrentCpuUsage()
		fmt.Println("cpu ", cpu)
		memory := cmetric.CurrentMemoryUsage()
		fmt.Println("memory ",memory/(1024 * 1024))

```

### test examples

```
func main() {

	go func() {
	restart:
		t := time.NewTicker(6 * time.Second)

		for {
			select {
			case <-t.C:
				fmt.Println("cpu down===========")
				time.Sleep(6 * time.Second)
				t.Stop()
				fmt.Println("cpu up===========")
				goto restart
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


```

i start a 0.5cpu container . then i exec the code in a container. we can see the cpu rate changes and the cpu is up to 0.5 the will not up to more.

```shell
root@019edf83f185:/go/src/cmetric/examples# ./examples 
2021/07/19 11:04:05 environment is  container
cpu  -1
memory  0
cpu  0.5143167487149263
memory  8
cpu  0.484085295650434
memory  8
cpu down===========
cpu  0.5033084169999711
memory  8
cpu  0.00460833034863899
memory  8
cpu  0.004765296517979556
memory  8
cpu up===========
cpu  0.004004104568513914
memory  8
cpu  0.5060104945249576
memory  8
cpu  0.4964300169136462
memory  9
cpu down===========
cpu  0.49390430746022085
memory  9
```


