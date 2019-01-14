package utils

import (
	"math/rand"
	"time"
)

const RANDOM_LETTERS = 10

// FlipCoin Flip coin
func FlipCoin() bool {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	// flipCoin
	coin := r.Int() % 2
	return coin != 0
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func CreateMinerHash() HashValue {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]rune, RANDOM_LETTERS)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return MakeHashString(string(b))
}
