package blockchain

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ageapps/gambercoin/pkg/utils"
)

func (bc *BlockChain) getBlockPool() map[string]*Block {
	bc.Lock()
	defer bc.Unlock()
	return bc.blockPool
}

func (bc *BlockChain) setPrevHash(newPrev [32]byte) {
	bc.Lock()
	bc.prevHash = newPrev
	bc.Unlock()
}

func (bc *BlockChain) setMining(yes bool) {
	bc.Lock()
	// if !yes {
	// 	logger.Logf("STOPPED MINIG")
	// }
	bc.minig = yes
	bc.Unlock()

}
func (bc *BlockChain) isMining() bool {
	bc.Lock()
	defer bc.Unlock()
	return bc.minig
}

func (bc *BlockChain) resetTransactionPool() {
	bc.Lock()
	bc.tansactionPool = make(map[string]*Transaction)
	bc.Unlock()
}

func (bc *BlockChain) setCurrentNonce(nonce [32]byte) {
	bc.Lock()
	bc.currentBlock.Nonce = nonce
	bc.Unlock()
}
func (bc *BlockChain) setCurrentBlock(block Block) {
	bc.Lock()
	bc.currentBlock = block
	bc.Unlock()
}
func (bc *BlockChain) restoreCanonicalChain(newChain Chain) {
	bc.Lock()
	bc.canonicalChain = newChain
	bc.Unlock()
}

func (bc *BlockChain) getCurrentBlock() Block {
	bc.Lock()
	defer bc.Unlock()
	return bc.currentBlock
}
func (bc *BlockChain) getCanonicalChain() Chain {
	bc.Lock()
	defer bc.Unlock()
	return bc.canonicalChain
}
func (bc *BlockChain) getPrevHash() utils.HashValue {
	bc.Lock()
	defer bc.Unlock()
	return bc.prevHash
}

func (bc *BlockChain) getTransactionPool() map[string]*Transaction {
	bc.Lock()
	defer bc.Unlock()
	return bc.tansactionPool
}

func (bc *BlockChain) getSideChains() []*Chain {
	bc.Lock()
	defer bc.Unlock()
	return bc.sideChains
}

func (bc *BlockChain) resetSideChains() {
	bc.Lock()
	bc.sideChains = []*Chain{}
	bc.Unlock()
}

func (bc *BlockChain) addSideChain(chain *Chain) {
	bc.Lock()
	bc.sideChains = append(bc.sideChains, chain)
	bc.Unlock()
}

func (bc *BlockChain) deleteSideChain(index int) {
	bc.sideChains = append(bc.sideChains[:index], bc.sideChains[index+1:]...)
}

func (bc *BlockChain) isTransactionInCanonicalChain(newTransaction *Transaction) bool {
	for _, block := range bc.getCanonicalChain().Blocks {
		for _, transaction := range block.Transactions {
			tx := transaction // because of ponter issues
			if tx.String() == newTransaction.String() {
				return true
			}
		}
	}
	return false
}

func (bc *BlockChain) logChain() {
	str := " "
	cChain := bc.getCanonicalChain()
	for index := cChain.size() - 1; index >= 0; index-- {
		block := cChain.Blocks[index]
		str += block.String()
		str += ":"
		str += hex.EncodeToString(block.PrevHash[:])
		str += ":"
		for index := 0; index < len(block.Transactions); index++ {
			str += block.Transactions[index].String()
			if index < len(block.Transactions)-1 {
				str += ","
			}
		}
		str += " "
	}
	fmt.Println("CHAIN" + str)
}

func (bc *BlockChain) isActive() bool {
	bc.Lock()
	defer bc.Unlock()
	return bc.active
}
func (bc *BlockChain) setActive(yes bool) {
	bc.Lock()
	bc.active = yes
	bc.Unlock()
}

func (bc *BlockChain) sendBlock(bl *Block) {
	if bc.isActive() {
		bc.sendChannel <- ChainMessage{
			Block:  bl,
			Origin: bc.nodeAddress,
		}
	}
}
func (bc *BlockChain) sendBlockToProcess(bl *Block) {
	if bc.isActive() {
		bc.ReceiveChannel <- ChainMessage{
			Block:  bl,
			Origin: bc.nodeAddress,
		}
	}
}
func (bc *BlockChain) sendTransaction(tx *Transaction) {
	if bc.isActive() {
		bc.sendChannel <- ChainMessage{
			Tx:     tx,
			Origin: bc.nodeAddress,
		}
	}
}
func (bc *BlockChain) getBlockTime() uint64 {
	bc.Lock()
	defer bc.Unlock()
	return bc.blockTime
}
func (bc *BlockChain) setBlockTime(t uint64) {
	bc.Lock()
	bc.blockTime = t
	bc.Unlock()
}

func getTimestamp() int64 {
	return time.Now().UnixNano()
}
