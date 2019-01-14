package utils

import (
	"math/rand"
	"time"
)

// FlipCoin Flip coin
func FlipCoin() bool {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	// flipCoin
	coin := r.Int() % 2
	return coin != 0
}
