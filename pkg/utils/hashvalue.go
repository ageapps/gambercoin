package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashValue is a file containing the SHA-256 hashes of each chunk
type HashValue [32]byte

// String method
func (hash *HashValue) String() string {
	bytes := hash[:]
	return hex.EncodeToString(bytes)
}

// Set HashValue from string
func (hash *HashValue) Set(value string) error {
	newHash, err := hex.DecodeString(value)
	for i := 0; i < 32; i++ {
		(*hash)[i] = newHash[i]
	}
	return err
}

// Equals from string
func (hash *HashValue) Equals(value string) bool {
	return hash.String() == value
}

// GetHash returns a HashValue
// from an string
func GetHash(value string) (HashValue, error) {
	var hash HashValue
	return hash, hash.Set(value)
}

// MakeHashString returns a HashValue
func MakeHashString(value string) HashValue {
	var hash HashValue = sha256.Sum256([]byte(value))
	return hash
	// hashArr := sha256.Sum256([]byte(value))
	// var hash HashValue
	// hash = hashArr[:]
	// return hash.String(), hash.Set(string(hash))
}
