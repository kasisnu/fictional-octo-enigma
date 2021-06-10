package main

import (
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

func main() {
	benchThis()
}

func benchThis() {
	num := 8
	var inputs []int
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	r := rand.New(rand.NewSource(99))
	for i := 0; i < num; i++ {
		inputs = append(inputs, r.Intn(100))
	}

	exitCh := make(chan string)
	returnCh := make(chan int)
	var cause string

	go func() {
		select {
		case <-time.After(10 * time.Second):
			cause = "timeout"
		case <-c:
			cause = "caught interrupt"
		}
		for i := 0; i < num; i++ {
			exitCh <- cause
		}
	}()

	// take my cpu - all of it
	for i := 0; i < num; i++ {
		go func(idx, val int) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Goodbye cause %s\n", r)
					returnCh <- 1
				}
			}()

			go cpuBound(val)

			reason := <-exitCh
			panic(reason)
		}(i, inputs[i])

	}
	for i := 0; i < num; i++ {
		<-returnCh
	}
}
