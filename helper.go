package main

import (
	"math/rand"
	"time"
)

// randomTimer returns a random duration between 5 and 10 seconds
func randomTimer() time.Duration {
	s := rand.Intn(6) + 5 // Random number between 5 and 10
	return time.Duration(s) * time.Second
}
