package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
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
	num := 100 // initial conditions
	poolSize := 2

	pool := NewPool(num, poolSize) // create a pool that listens for signals

	// create work
	go func() {
		for i := 0; i < num; i++ {
			pool.WorkCh <- func() { // send task down to pool
				timeBound(10)
			}
		}
	}()

	pool.Wait() // Wait for all tasks to finish
}

// Pool of tasks of type func() where workload size is known ahead of time
// todo(kasisnu): rewrite for streaming workloads
type Pool struct {
	WorkCh chan func() // Queues work or blocks if pool is at capacity

	poolSize     *uint32        // current size of pool
	numActive    int            // current active workers
	numCompleted int            // current count of work that')s completed
	workloadSize int            // could be removed later by introducing a signal to free explicitly
	sigChan      chan os.Signal // channel for pool size signalling
	returnCh     chan int
}

func NewPool(workloadSize, poolSize int) Pool {
	p := Pool{
		numActive:    0,
		numCompleted: 0,
		workloadSize: workloadSize,
	}

	poolSizeuint32 := uint32(poolSize)
	p.poolSize = &poolSizeuint32

	p.WorkCh = make(chan func())
	p.returnCh = make(chan int, 2)
	p.sigChan = make(chan os.Signal, 2)
	go p.processSignals() // start listening for signals
	return p
}

func (p *Pool) WorkloadSize() int {
	return p.workloadSize
}

func (p *Pool) Wait() { // wait for work and close channels on completion
	p.processWork()
	p.done()
}

func (p *Pool) done() {
	close(p.WorkCh)
	close(p.sigChan)
}

func (p *Pool) processSignals() {
	signal.Notify(p.sigChan,
		syscall.SIGKILL,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGUSR1,
		syscall.SIGUSR2)

	for interrupt := range p.sigChan {
		switch interrupt {
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGQUIT:
			os.Exit(1)
		case syscall.SIGUSR1:
			p.setPoolSizeDeltaSafe(+1)
			fmt.Printf("Pool size increased by 1. Current Pool size: %d\n", p.currentPoolSize())
		case syscall.SIGUSR2:
			p.setPoolSizeDeltaSafe(-1)
			fmt.Printf("Pool size decreased by 1. Current Pool size: %d\n", p.currentPoolSize())
		}
	}
}

func (p *Pool) ackOneTaskFinished() {
	p.numActive = p.numActive - 1
	p.numCompleted = p.numCompleted + 1
}

func (p *Pool) processWork() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var expectedPoolSize int

	for { //synchronous loop for internal state mutations
		expectedPoolSize = p.currentPoolSize()

		if p.numCompleted == p.workloadSize {
			return
		}

		if p.numActive >= (p.workloadSize - p.numCompleted) { // There aren't any additional tasks left
			<-p.returnCh // there's at least one task running,
			//and there isn't more work left so let's wait for one task to return, we'll be back here if there are more left
			p.ackOneTaskFinished()
			continue // skip trying to start new work cause we've exhausted workloads
		}

		task := <-p.WorkCh // a new task is ready to run and
		// we're blocked because pool is full
		for p.numActive >= expectedPoolSize {
			select {
			case <-p.returnCh: // wait for a task to finish
				p.ackOneTaskFinished() // notify that one task finished
			case <-ticker.C: //pool size might have changed since we started waiting, check again
				expectedPoolSize = p.currentPoolSize()
				continue
			}
		}

		p.numActive = p.numActive + 1 // update task counter

		go func() { // if we're here, there's capacity for at least one task
			task()          // start task and notify self when finished
			p.returnCh <- 1 // notify self that task has finished // todo(kasisnu): this should be in recover
		}()

		// print some debugging info -- todo(kasisnu): put behind flag
		fmt.Println("active/completed/pending/poolsize", p.numActive, p.numCompleted, p.workloadSize-p.numCompleted-p.numActive, p.currentPoolSize())
	}
}

func (p *Pool) currentPoolSize() int {
	return int(atomic.LoadUint32(p.poolSize))
}

func (p *Pool) setPoolSizeDeltaSafe(delta int) {
	oldPoolSizeAddr := p.poolSize
	oldPoolSize := atomic.LoadUint32(oldPoolSizeAddr)
	if oldPoolSize <= 0 {
		return
	}
	deltaUint32 := uint32(delta)
	newPoolSizeAddr := atomic.AddUint32(p.poolSize, deltaUint32)
	atomic.SwapUint32(oldPoolSizeAddr, newPoolSizeAddr)
}
