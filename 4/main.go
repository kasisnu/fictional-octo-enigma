package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
)

func init() {
	runtime.SetCPUProfileRate(500)
}

func cpuBound(ctx context.Context, n int) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	//occasionally check if we should return
	ticker := time.NewTicker(time.Millisecond * 10)
	for {
		fmt.Fprintf(f, ".")
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			continue
		}
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
	num := 16
	var inputs []int
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	r := rand.New(rand.NewSource(99))
	for i := 0; i < num; i++ {
		inputs = append(inputs, r.Intn(100))
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		cancel()
	}()

	returnCh := make(chan int)
	// take my cpu - all of it
	for i := 0; i < num; i++ {
		go func(idx, val int) {
			defer fmt.Printf("Goodbye %d\n", idx)
			cpuBound(ctx, val)
			returnCh <- 1
		}(i, inputs[i])
	}

	for i := 0; i < num; i++ {
		<-returnCh
	}
}
