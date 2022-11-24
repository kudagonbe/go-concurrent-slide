package dice

import (
	"fmt"
	"math/rand"
	"time"
)

func RollDice(index int) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(6) + 1
	fmt.Printf("%d: dice number is %d\n", index, r)
}

func RollDiceWithChannel(ch chan int) {
	rand.Seed(time.Now().UnixNano())
	ch <- rand.Intn(6) + 1
}
