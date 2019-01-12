package blockchain

import (
	"crypto/sha256"

	"github.com/ageapps/gambercoin/pkg/utils"
)

// Transaction struct
type Transaction struct {
	Input  utils.HashValue
	Output utils.HashValue
	Name   utils.HashValue
}

// AppendTransaction func
func (tx *Transaction) String() string {
	return tx.Name.String()
}

// Hash transaction
func (tx *Transaction) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(tx.Name[:])
	h.Write(tx.Input[:])
	h.Write(tx.Output[:])
	copy(out[:], h.Sum(nil))
	return
}
