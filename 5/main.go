package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

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
	benchThis()
}

func benchThis() {
	num := 10
	var poolSize = flag.Int("poolsize", num, "number of worker goroutines per pool")
	flag.Parse()

	var numActive int
	var numCompleted int

	workCh := make(chan func())
	returnCh := make(chan int)

	go func() {
		for i := 0; i < num; i++ {

			workCh <- func() {
				timeBound(10)
				returnCh <- 1
			}

		}
		close(workCh)
	}()

	for f := range workCh {
		numActive = numActive + 1
		go f()
		if numActive >= *poolSize {
			<-returnCh
			numCompleted = numCompleted + 1
			numActive = numActive - 1
		}
		time.Sleep(50 * time.Millisecond)
	}
}
