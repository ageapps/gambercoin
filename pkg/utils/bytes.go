package utils

import (
	"encoding/hex"
)

// Bytes array on steroids
type Bytes []byte

// String method
func (hash *Bytes) String() string {
	return hex.EncodeToString(*hash)
}

// Set HashValue from string
func (hash *Bytes) Set(value string) error {
	newHash, err := hex.DecodeString(value)
	*hash = newHash
	return err
}

// Equals from string
func (hash *Bytes) Equals(value string) bool {
	return hash.String() == value
}
