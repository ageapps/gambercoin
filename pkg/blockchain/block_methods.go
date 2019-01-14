package blockchain

import (
	"encoding/hex"

	"github.com/ageapps/gambercoin/pkg/logger"
)

// isBlockValid to the blockchain
func (bc *BlockChain) isBlockValid(bl *Block) bool {
	if !checkZeros(bl.Nonce) {
		return false
	}
	return !(bc.getBlockType(bl) == BLOCK_OLD)
}

// checkZeros
func checkZeros(nonce [32]byte) bool {
	prefix := nonce[0:NumberOfZeros]
	flag := true
	for _, num := range prefix {
		flag = flag && int(num) == 0
	}
	return flag
}

func (bc *BlockChain) getBlockType(newBlock *Block) string {
	// same prev hash than out prev block
	canonicalChain := bc.getCanonicalChain()
	if canonicalChain.isNextBlockInChain(newBlock) {
		return BLOCK_CURRENT
	}
	_, ok := bc.getBlockPool()[newBlock.String()]
	if ok {
		logger.Logv("Block already in pool")
		return BLOCK_OLD
	}

	if existing := bc.isBlockInCanonicalChain(newBlock); existing {
		logger.Logv("Block already in chain")
		return BLOCK_OLD
	}

	return BLOCK_UNKOWN_PARENT
}

func (bc *BlockChain) isBlockInCanonicalChain(newBlock *Block) bool {
	for _, block := range bc.getCanonicalChain().Blocks {
		if block.String() == newBlock.String() {
			return true
		}
	}
	return false
}

func (bc *BlockChain) addBlock(newBlock *Block, forking bool) (added bool) {
	if !checkZeros(newBlock.Nonce) {
		return false
	}
	blockType := bc.getBlockType(newBlock)
	logger.Logb("Adding Block of type %v", blockType)

	switch blockType {

	case BLOCK_CURRENT:
		// stop mining
		bc.setMining(false)
		bc.addToBlockChain(newBlock)
		if !bc.checkTransactionsInCurrentBlock(newBlock) {
			logger.Logf("Transactions NOT found in current block")
			// there is no transactions in the block that i am currently minig,
			// jclean transaction pool from new block and keep mining
			bc.cleanTransactionPoolByAddedBlock(newBlock)
			if !forking {
				bc.buildBlockAndMine()
			}
		} else {
			logger.Logf("Transactions found in current block")
			// fmt.Println(newBlock.Transactions)
			// fmt.Println(bc.getCurrentBlock().Transactions)

			// there is transactions in the currentBlock,
			// add them again to the Txpool, clean transaction pool from new block,
			// reset current block and rebuild it to mine again
			bc.addBlockTransactionsToPool(bc.getCurrentBlock())
			bc.cleanTransactionPoolByAddedBlock(newBlock)
			bc.resetCurrentBlock()
			if !forking {
				bc.buildBlockAndMine()
			}
		}

		return true
	case BLOCK_UNKOWN_PARENT:
		if forking {
			logger.Logw("SOMETHING IS WRONG, UNKOWN PARENT FOUND WHILE FORKING")
		}
		// just save it
		sideChainIndex, parentIndex := bc.addToBlockPool(newBlock)
		if sideChainIndex >= 0 && parentIndex >= 0 {
			logger.Logf("Side chains found")
			bc.setMining(false)
			bc.forkCanonicalChain(sideChainIndex, parentIndex)
			bc.buildBlockAndMine()
		} else {
			logger.Logf("NO Side chains found")
		}

		return true
	case BLOCK_OLD:
		logger.Logf("Block already in blockchain, dropping it...")
		// if it's an old newBlock, just discard it
		return false
	}
	return false
}

// addToBlockPool adds block to pool
// and return references to a big sideChain
func (bc *BlockChain) addToBlockPool(block *Block) (sideChainIndex, parentIndex int) {
	logger.Logf("Adding Block to Blok Pool - %v", block.String())
	bc.Lock()
	bc.blockPool[block.String()] = block
	bc.Unlock()
	// everytime a block is added to the block pool
	// explore and build sidechains
	bc.buildSideChains(block)
	return bc.checkLongestChain()
}

func (bc *BlockChain) findNextBlock(block *Block) *Block {
	for _, newBlock := range bc.getBlockPool() {
		if hex.EncodeToString(newBlock.PrevHash[:]) == hex.EncodeToString(block.Nonce[:]) {
			return newBlock
		}
	}
	return nil
}
