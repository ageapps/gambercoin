package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// Block stuct
type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []Transaction
	timestamp    string
	TXCount      int
}

// NewBlock func
func NewBlock(prev [32]byte) *Block {
	return &Block{
		PrevHash:     prev,
		Transactions: []Transaction{},
		// timestamp: // TODO create timestamp
		TXCount: 0,
	}
}

// AppendTransaction func
func (block *Block) AppendTransaction(tx Transaction) {
	block.Transactions = append(block.Transactions, tx)
	block.TXCount = len(block.Transactions)
}

// AppendTransaction func
func (block *Block) String() string {
	return hex.EncodeToString(block.Nonce[:])
}

// PrintPrev func
func (block *Block) PrintPrev() string {
	return hex.EncodeToString(block.PrevHash[:])
}

// IsNextBlock func
func (block *Block) IsNextBlock(newBlock *Block) bool {
	if hex.EncodeToString(newBlock.PrevHash[:]) == hex.EncodeToString(block.Nonce[:]) {
		return true
	}
	return false
}

// Hash block
func (block *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(block.PrevHash[:])
	h.Write(block.Nonce[:])
	h.Write([]byte(block.timestamp))
	binary.Write(h, binary.LittleEndian, uint32(len(block.Transactions)))
	for _, t := range block.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}
