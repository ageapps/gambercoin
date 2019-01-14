package blockchain

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/ageapps/gambercoin/pkg/utils"
)

// Transaction struct
type Transaction struct {
	Input  utils.HashValue
	Output utils.HashValue
	Name   utils.HashValue
	Amount uint32
}

func NewTransaction(in, out utils.HashValue, amount int) Transaction {
	tx := Transaction{
		Input:  in,
		Output: out,
		Amount: uint32(amount),
	}
	tx.Name = tx.Hash()
	return tx
}

// AppendTransaction func
func (tx *Transaction) String() string {
	return tx.Name.String()
}

// Hash transaction
func (tx *Transaction) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, tx.Amount)
	h.Write(tx.Input[:])
	h.Write(tx.Output[:])
	copy(out[:], h.Sum(nil))
	return
}
