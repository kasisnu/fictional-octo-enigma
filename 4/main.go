package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
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

// genrandIntsList returns a slice of ints
// of length and in range (min, max-1)
func genrandIntsList(min, max, length int) []int {
	xs := make([]int, length)

	r := rand.New(rand.NewSource(99))
	for i := 0; i < length; i++ {
		xs[i] = r.Intn(max-min) + min
	}
	return xs
}

func main() {
	benchThis()
}

func benchThis() {
	num := 4
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // listen for os interrupts

	returnCh := make(chan int)                              // notifies task runner about task completion
	ctx, cancel := context.WithCancel(context.Background()) // create cancellable context
	go func() {                                             // notify tasks about shutdown
		select {
		case <-c: // check for any os signals
		case <-time.After(10 * time.Second): // check if we should timeout
		}
		cancel() // send cancel down the context
	}()

	//generate inputs
	inputs := genrandIntsList(0, 100, num)
	for i := 0; i < num; i++ {
		go func(ctx context.Context, idx, val int) {
			defer func() {
				if r := recover(); r != nil { // ensure we always notify task runner

					fmt.Printf("Goodbye %d\n", idx)

					returnCh <- 1 // propagate shutdown back from this routine
				}
			}()

			go cpuBound(val)

			<-ctx.Done()      //task is running now, we'll w ait for a shutdown signal(via cancel)
			panic("stopping") // unwind stack to ensure we notify about completion
		}(ctx, i, inputs[i])
	}

	// wait for everything to finish
	for i := 0; i < num; i++ {
		<-returnCh
	}
}
