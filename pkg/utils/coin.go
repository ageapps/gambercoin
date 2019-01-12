package utils

import (
	"math/rand"
	"time"
)

// KeepRumorering Flip coin
func KeepRumorering() bool {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	// flipCoin
	coin := r.Int() % 2
	return coin != 0
}
