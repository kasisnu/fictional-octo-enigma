package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"
)

func init() {
	runtime.SetCPUProfileRate(500)
}

func timeBound(n int) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for x := 0; x < n; x++ {
		fmt.Fprintf(f, ".")
		time.Sleep(1 * time.Second)
	}
}

func main() {
	var cpuProfile = flag.String("cpuprofile", "cpu.pprof", "write cpu profile to file")
	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)

		benchThis()
	}
}

func benchThis() {
	num := 10
	poolSize := 2

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	returnCh := make(chan int)

	// poor man's pooler
	var numActive int
	var numCompleted int
	var mu sync.Mutex
	for numCompleted != num {
		fmt.Println(numActive, numCompleted)
		go func(mu *sync.Mutex) {
			func() {
				mu.Lock()
				defer mu.Unlock()
				numActive = numActive + 1
			}()
			timeBound(10)
			fmt.Println("stopped one")
			returnCh <- 1
		}(&mu)
		if numActive >= poolSize {
			<-returnCh
			func() {
				mu.Lock()
				defer mu.Unlock()
				numActive = numActive - 1
			}()
			numCompleted = numCompleted + 1
		}
		time.Sleep(100 * time.Millisecond)
	}
}
