package main

import (
	"fmt"
	"os"
	"os/signal"
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
		os.Exit(0)
	}

	fmt.Println("Goodbye")
}

func main() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // listen for os interrupts

	// try to handle os interrupt(signal terminated)
	go onKill(c)

	benchThis()
}

func benchThis() {
	cpuBound(0)
}
