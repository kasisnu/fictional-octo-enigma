package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
)

func cpuBound(n int) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for {
		fmt.Fprintf(f, ".")
	}
}

func onKill(c chan os.Signal) {
	select {
	case <-c:
		defer pprof.StopCPUProfile()
		defer os.Exit(0)
	}

	fmt.Println("Goodbye")
}

func main() {
	runtime.SetCPUProfileRate(500)
	var cpuProfile = flag.String("cpuprofile", "cpu.pprof", "write cpu profile to file")
	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)

		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM) // subscribe to system signals

		// try to handle os interrupt(signal terminated)
		go onKill(c)

		benchThis()
	}
}

func benchThis() {
	cpuBound(0)
}
