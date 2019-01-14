package blockchain

import (
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/utils"
)

// isTransactionValid func
// check if is in transaction pool or in the canonical chain
func (bc *BlockChain) isTransactionValid(tx *Transaction) bool {
	_, ok := bc.getTransactionPool()[tx.String()]
	if ok {
		return true
	}
	return bc.isTransactionInCanonicalChain(tx)
}

func (bc *BlockChain) addToTransactionPool(tx *Transaction) utils.HashValue {
	logger.Logb("Storing - %v in TXpool", tx.String())
	bc.Lock()
	bc.tansactionPool[tx.String()] = tx
	bc.Unlock()
	return tx.Name
}

func (bc *BlockChain) deleteFromTxPool(hash string) {
	bc.Lock()
	delete(bc.tansactionPool, hash)
	bc.Unlock()
}

func (bc *BlockChain) cleanTransactionPoolByAddedBlock(newBlock *Block) {
	// Look in tx pool
	for _, newTx := range newBlock.Transactions {
		bc.deleteFromTxPool(newTx.String())
	}
}
