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

