package blockchain

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"

	"github.com/ageapps/gambercoin/pkg/utils"
)

// TransactionMulti struct
type TransactionMulti struct {
	Input  Input
	Output Output
	NumIn  int
	NumOut int
	Name   utils.HashValue
}

// VerifySignature verifies the integrity if the transaction
func (tx *TransactionMulti) VerifySignature() (err error) {
	return rsa.VerifyPKCS1v15(tx.Input.PubKey, crypto.SHA256, tx.Name[:], tx.Input.Signature)
}

// AppendTransaction func
func (tx *TransactionMulti) String() string {
	return tx.Name.String()
}

// Input struct
// PrevOut reference to previous output
// Index index of output in referenced transaction
type Input struct {
	PrevOut   utils.HashValue
	Index     int
	Signature utils.Bytes
	PubKey    *rsa.PublicKey
}

// Hash input
func (in *Input) Hash() (out []byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(in.Index))
	h.Write(in.Signature)
	copy(out[:], h.Sum(nil))
	return
}

// Output struct
// PubKeyHash hash of receivers publick key
// Value number of coins sended
type Output struct {
	PubKeyHash utils.HashValue
	Value      int
}

// Hash transaction
func (tx *TransactionMulti) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(tx.NumIn))
	binary.Write(h, binary.LittleEndian, uint32(tx.NumOut))
	h.Write(tx.Name[:])
	h.Write(tx.Input.Hash())
	h.Write(tx.Output.PubKeyHash[:])
	copy(out[:], h.Sum(nil))
	return
}
