package pangolin

import (
  "io/ioutil"
  "log"
  "os"
  "os/signal"
  "path/filepath"
  "runtime"
  "runtime/pprof"
  "sync/atomic"

)

type modeType int

const (
  CPUMODE modeType = iota
  MEMMODE
  BLKMODE
)

var (
  started int32
)

type profile struct{
  rate int
  mode modeType
  path string
  noInterruptHook bool
  stopped int32
  stopper func()
}

func (prof profile) Stop() {
  if atomic.CompareAndSwapInt32(&prof.stopped, 0, 1) {
    prof.stopper()
    atomic.StoreInt32(&started, 0)
  }
}

// profile decorators
func CPUProfile(prof profile) profile {
  prof.mode = CPUMODE
  return prof
}

func MemProfile(prof profile) profile {
  prof.rate = runtime.MemProfileRate
  prof.mode = MEMMODE
  return prof
}

func MemProfileRate(r int) func(profile) profile {
  f := func(prof profile) profile {
    prof.rate = r
    prof.mode = MEMMODE
    return prof
  }
  return f
}

func BlockProfile(prof profile) profile {
  prof.mode = BLKMODE
  return prof
}

func NoInterruptHook() func(profile) profile {
  f := func(prof profile) profile {
    prof.noInterruptHook = true
    return prof
  }
  return f
}

func ProfilePath(path string) func(profile) profile {
  f := func(prof profile) profile {
    prof.path = path
    return prof
  }
  return f
}

func Start(decorators ...func(profile) profile) interface {
  Stop()
} {
  if atomic.CompareAndSwapInt32(&started, 0, 1) {
    prof := profile{}
    for _, dec := range decorators {
      prof = dec(prof)
    }

    profDir, err := func() (string, error) {
      if prof.path == "" {
        return ioutil.TempDir("", "profile")
      } else {
        return prof.path, os.MkdirAll(prof.path, os.ModePerm)
      }
    }()

    if err != nil {
      log.Fatalln(err)
    }

    switch prof.mode {
    case CPUMODE:
      profName := filepath.Join(profDir, "cpu.prof")
      f, err := os.Create(profName)
      if err != nil {
        log.Fatalln(err)
      }
      if err = pprof.StartCPUProfile(f); err != nil {
        f.Close()
        log.Fatalln(err)
      }
      prof.stopper = func() {
        pprof.StopCPUProfile()
        f.Close()
      }

    case MEMMODE:
      profName := filepath.Join(profDir, "mem.prof")
      f, err := os.Create(profName)
      if err != nil {
        log.Fatalln(err)
      }
      oldRate := runtime.MemProfileRate
      runtime.MemProfileRate = prof.rate
      prof.stopper = func() {
        pprof.Lookup("heap").WriteTo(f, 0)
        f.Close()
        runtime.MemProfileRate = oldRate
      }

    case BLKMODE:
      profName := filepath.Join(profDir, "blk.prof")
      f, err := os.Create(profName)
      if err != nil {
        log.Fatalln(err)
      }
      runtime.SetBlockProfileRate(1)
      prof.stopper = func() {
        pprof.Lookup("block").WriteTo(f, 0)
        f.Close()
        runtime.SetBlockProfileRate(0)
      }

    }

    if !prof.noInterruptHook {
      go func() {
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt)
        <-c

        prof.Stop()
        os.Exit(0)
      }()
    }

    return prof
  }
  panic("Start() already called")
}
