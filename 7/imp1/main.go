package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func mathBound(n int) int {
	time.Sleep(time.Duration(n) * time.Second)
	return n * 8
}

func genrandIntsList(min, max, length int) []int {
	xs := make([]int, length)

	for i := 0; i < length; i++ {
		xs[i] = rand.Intn(max-min) + min
	}
	return xs
}

func processWithMutex(inputs []int) map[int]int {
	var mu sync.Mutex
	var wg sync.WaitGroup

	outputs := make(map[int]int)

	for i, v := range inputs {
		wg.Add(1)
		go func(mu *sync.Mutex, wg *sync.WaitGroup, index, duration int) {
			defer wg.Done()

			res := mathBound(duration)

			mu.Lock()
			defer mu.Unlock()
			outputs[index] = res

		}(&mu, &wg, i, v)
	}

	wg.Wait()
	return outputs
}

func main() {
	inputs := genrandIntsList(0, 10, 5)
	fmt.Println(inputs)

	fmt.Println(processWithMutex(inputs))

}
