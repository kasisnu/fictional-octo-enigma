package main

import (
	"fmt"
	"math/rand"
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

func processWithoutMutex(inputs []int) map[int]int {
	outputs := make(map[int]int)
	resultChan := make(chan [2]int)

	for i, v := range inputs {
		go func(resultChan chan [2]int, index int, duration int) {
			res := mathBound(duration)
			resultChan <- [2]int{index, res}
		}(resultChan, i, v)
	}

	for i := 0; i < len(inputs); i++ {
		inputResultPair := <-resultChan
		outputs[inputResultPair[0]] = inputResultPair[1]
	}

	return outputs
}

func main() {
	inputs := genrandIntsList(0, 10, 5)
	fmt.Println(processWithoutMutex(inputs))
}
