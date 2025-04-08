package state

import (
	"fmt"
	"math/rand"
	"time"
)

// randomTimer returns a random duration between 5 and 10 seconds
func RandomTimer() time.Duration {
	s := rand.Intn(15) + 15 // Random number between 5 and 10
	return time.Duration(s) * time.Second
}

func Countdown(n *Node) {
	elapsed := 0
	for elapsed < int(n.Timer.Seconds()) {
		fmt.Println("elapsed time:", elapsed)
		time.Sleep(1 * time.Second)
		elapsed++
	}
}
