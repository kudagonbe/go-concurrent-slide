package main

import (
	"fmt"

	"github.com/kudagonbe/go-concurrent-slide/dice"
)

func main() {
	// call RollDice 10 times
	// for i := 0; i < 10; i++ {
	// 	dice.RollDice(i)
	// }

	// call RollDice 10 times with goroutine
	// for i := 0; i < 10; i++ {
	// 	go dice.RollDice(i)
	// }

	// call RollDice 10 times with goroutine and waitgroup
	// var wg sync.WaitGroup
	// for i := 0; i < 10; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		dice.RollDice(i)
	// 	}(i)
	// }
	// wg.Wait()

	var count int = 100000
	var total int
	values := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0}

	ch := make(chan int, count)
	done := make(chan struct{})

	for i := 0; i < count; i++ {
		go dice.RollDiceWithChannel(ch)
	}

Loop:
	for {
		select {
		case value := <-ch:
			total++
			values[value]++
			if total >= count {
				close(done)
			}
		case <-done:
			close(ch)
			break Loop
		}
	}

	outputResults(values, total)
}

func outputResults(values map[int]int, total int) {
	fmt.Printf("total: %d\n", total)
	fmt.Printf("1: %d\n", values[1])
	fmt.Printf("2: %d\n", values[2])
	fmt.Printf("3: %d\n", values[3])
	fmt.Printf("4: %d\n", values[4])
	fmt.Printf("5: %d\n", values[5])
	fmt.Printf("6: %d\n", values[6])
	fmt.Printf("average: %f\n", avg(values, total))

}

func avg(values map[int]int, total int) float64 {
	var f float64
	for k, v := range values {
		f += float64(k) * float64(v)
	}
	return f / float64(total)
}
