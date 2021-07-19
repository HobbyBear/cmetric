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

```

i start a 0.5cpu container . then i exec the code in a container. we can see the cpu rate changes and the cpu is up to 0.5 the will not up to more.

```shell
root@019edf83f185:/go/src/cmetric/examples# ./examples 
2021/07/19 06:01:19 environment is  container
cpu  -1
memory  0
cpu  0.49642350447547673
memory  13
cpu  0.49847000400040997
memory  13
cpu down
cpu  0.49710812400007853
memory  13
cpu  0.004654334000406379
memory  13
cpu  0.003783542713734267
memory  13
cpu down
cpu  0.004830249999940861
memory  13
cpu  0.004333666331600389
memory  13
cpu  0.004318668000450998
memory  14
cpu  0.0039140181815438815
memory  14
cpu  0.502048414072731
memory  14
cpu  0.5042367737423622
memory  14
cpu down
cpu  0.5057867512758194
memory  14
cpu  0.00489279399971565
memory  14
cpu  0.0035706422114973315
memory  14

```


