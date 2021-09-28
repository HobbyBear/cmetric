## read metrics in container


if environment is container, the cpu ,memory is relative to container,
else  the metrics is relative to host.

juejing link : https://juejin.cn/post/6986598285406371871/

### usage

```

cpu := cmetric.CurrentCpuPercentUsage()
fmt.Println("cpu ", cpu)
memory := cmetric.CurrentMemoryPercentUsage()
fmt.Println("memory ", memory)
time.Sleep(2000 * time.Millisecond)

```

### test examples

```
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

```

i start a 0.5cpu container . then i exec the code in a container. we can see the cpu rate changes and the cpu is up to 0.5 the will not up to more.

```shell
pu  0
memory  0.0009067918
cpu  1.0382111932603482
memory  0.00097522896
cpu  1.0270122977276814
memory  0.00097522896
cpu  0.9752556716859127
memory  0.00097522896
cpu down===========
cpu  0.01998823174866435
memory  0.00097522896
cpu  0
memory  0.00097522896
cpu  0
memory  0.00097522896
cpu up===========
cpu  1.0202506439554495
memory  0.0010586367
cpu  1.027026345137049
memory  0.0010586367
cpu  1.0184514486461407
memory  0.0010586367
cpu down===========
cpu  0
memory  0.0010586367
cpu  0
memory  0.0010586367
```


