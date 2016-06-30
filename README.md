pangolin
=======

Golang server profiling package


installation
------------

    go get github.com/leesper/pangolin

usage
-----

Add one line at the top of your main function, and profiling will be enabled during server running. Then you can analysis the .prof file by go tool pprof


```go
import "github.com/leesper/pangolin"

func main() {
    defer pangolin.Start().Stop()
    ...
}
```

decorators
-------
* CPUProfile
* MemProfile
* MemProfileRate(r int)
* BlockProfile
* NoInterruptHook
* ProfilePath(path string)

The CPU profiling mode is enabled by default, and you can alter behaviours such as changing profile directory or enabling memory profiling by passing decorators into pangolin.Start()

```go
import "github.com/leesper/pangolin"

func main() {
    defer pangolin.Start(pangolin.MemProfile, pangolin.ProfilePath("./prof"), pangolin.NoShutdownHook).Stop()
    ...
}
```
