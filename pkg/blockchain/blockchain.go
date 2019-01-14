package blockchain

import (
	"sync"

	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/utils"
)

const (
	// NumberOfZeros in the nonce
	NumberOfZeros = 2
	// BLOCK_OLD const
	BLOCK_OLD = "BLOCK_OLD"
	// BLOCK_CURRENT const
	BLOCK_CURRENT = "BLOCK_CURRENT"
	// BLOCK_UNKOWN_PARENT const
	BLOCK_UNKOWN_PARENT = "BLOCK_NEW"
)

// BlockChain struct
type BlockChain struct {
	sendChannel    chan<- ChainMessage //  channel to send messages to node
	ReceiveChannel chan ChainMessage   // write-only channel to receive messages from node

	minig       bool
	active      bool //   monguer handler active state
	blockTime   uint64
	nodeAddress string
	minerHash   utils.HashValue

	canonicalChain Chain
	sideChains     []*Chain
	currentBlock   Block
	prevHash       utils.HashValue

	tansactionPool map[string]*Transaction
	blockPool      map[string]*Block

	quitChannel chan bool

	sync.Mutex
}

// NewBlockChain func
func NewBlockChain(nodeAddress string, minerHash utils.HashValue) *BlockChain {
	return &BlockChain{
		minig:       false,
		active:      false,
		nodeAddress: nodeAddress,
		minerHash:   minerHash,

		canonicalChain: NewEmptyChain(),
		sideChains:     []*Chain{},
		prevHash:       [32]byte{},

		tansactionPool: make(map[string]*Transaction),
		blockPool:      make(map[string]*Block),

		quitChannel: make(chan bool),
	}

}

// Start blockchain process
func (bc *BlockChain) Start(onStopHandler func()) <-chan ChainMessage {
	bc.sendChannel = make(chan ChainMessage)
	bc.ReceiveChannel = make(chan ChainMessage)
	go func() {
		for {
			select {
			case message := <-bc.ReceiveChannel:
				if message.IsTx() {
					tx := message.Tx
					bc.processTransaction(tx)
				} else if message.IsBlock() {
					bl := message.Block
					bc.processBlock(bl)
				} else {
					logger.Logw("Message received not recognized")
				}

			case <-bc.quitChannel:
				bc.setMining(false)
				logger.Logf("Finishing Blockchain")
				close(bc.sendChannel)
				onStopHandler()
				return
			}
		}
	}()
	return bc.ReceiveChannel
}

func (bc *BlockChain) processTransaction(tx *Transaction) {
	if !bc.isTransactionValid(tx) {
		return
	}
	bc.addToTransactionPool(tx)
	if !bc.isMining() {
		bc.buildBlockAndMine()
	}
}

func (bc *BlockChain) processBlock(bl *Block) {
	if !bc.isBlockValid(bl) {
		return
	}
	bc.addBlock(bl, false)
}

func (bc *BlockChain) buildBlockAndMine() {
	canonicalChain := bc.getCanonicalChain()
	for _, block := range bc.getBlockPool() {
		if canonicalChain.isNextBlockInChain(block) {
			logger.Logf("Found stored block matching prev - %v", block.String())
			bc.processBlock(block)
			return
		}
	}
	// Check mining and transactions available
	if len(bc.getTransactionPool()) > 0 && !bc.isMining() {
		// -> Build new block from prevhash
		currentBlock := *NewBlock(bc.getPrevHash())
		// -> Add coinbase to block
		coinBase := Transaction{
			Input:  [32]byte{},
			Output: bc.minerHash,
			Amount: uint32(1),
		}
		currentBlock.AppendTransaction(coinBase)
		// -> Fill block with transactions from pool
		for _, tx := range bc.getTransactionPool() {
			currentBlock.AppendTransaction(*tx)
		}
		// -> Set as Current block
		bc.setCurrentBlock(currentBlock)
		logger.Logf("Mining current block with transactions %v", len(currentBlock.Transactions))
		// Mine it
		go func() {
			bc.mine()
		}()
	} else {
		logger.Logf("No transactions to mine/ already minig...")
	}
}

func (bc *BlockChain) mine() {
	bc.setMining(true)
	currentBlock := bc.getCurrentBlock()
	prev := bc.getPrevHash()
	logger.Logf("Expecting - %v", prev.String())
	logger.Logf("Mining block with parent - %v", currentBlock.PrintPrev())
	init := getTimestamp()
	for bc.isMining() {
		bc.resetCurrentBlock()
		nonce := currentBlock.Hash()
		currentBlock.Nonce = nonce
		if checkZeros(nonce) {
			bc.setBlockTime(uint64(getTimestamp() - init))
			logger.LogFoundBlock(currentBlock.String())
			// Send block to main routine to process it
			bc.sendBlockToProcess(&currentBlock)
			break
		}
	}
	logger.Logf("FINISHED MINIG")
	bc.setMining(false)
}

func (bc *BlockChain) addBlockTransactionsToPool(newBlock Block) {
	// Look in tx pool
	for _, newTx := range newBlock.Transactions {
		bc.addToTransactionPool(&newTx)
	}
}

func (bc *BlockChain) resetCurrentBlock() {
	bc.setCurrentBlock(*NewBlock(bc.getPrevHash()))
}

// check if new block contains transactions that i'm currently mining
func (bc *BlockChain) checkTransactionsInCurrentBlock(newBlock *Block) bool {
	cb := bc.getCurrentBlock()
	// Look in current block and non coincident tx, add to tx pool
	for _, newTx := range newBlock.Transactions {
		for _, tx := range cb.Transactions {
			if newTx.String() == tx.String() {
				return true
			}
		}
	}
	return false
}

func (bc *BlockChain) addToBlockChain(block *Block) {
	// reference the prev hash to the new added block
	bc.setPrevHash(block.Nonce)

	bc.Lock()
	logger.Logf("Adding Block to Canonical Chain - %v", block.String())
	logger.Logf("With prev - %v", block.PrintPrev())
	bc.canonicalChain.appendBlock(block)
	bc.Unlock()
	bc.logChain()
}

// build all sidechains possible
func (bc *BlockChain) buildSideChains(block *Block) {
	bc.resetSideChains()
	logger.Logf("Trying to build new sidechains...")
	for _, block := range bc.getBlockPool() {
		newChain := NewEmptyChain()
		lastBlock := block
		for lastBlock != nil {
			newChain.appendBlock(lastBlock)
			lastBlock = bc.findNextBlock(lastBlock)
		}
		if newChain.size() > 1 {
			logger.Logf("Found sidechain of size - %v", newChain.size())
			bc.addSideChain(&newChain)
		}
	}
}

func (bc *BlockChain) forkCanonicalChain(sideChainIndex, parentIndex int) {
	canonicalChain := bc.getCanonicalChain()
	headCanonicalChain := canonicalChain.getSubchain(0, parentIndex+1)
	removingChain := canonicalChain.getSubchain(parentIndex+1, canonicalChain.size())
	logger.LogForkLong(len(removingChain.Blocks))

	// add removed block's transactions to pool
	for _, block := range removingChain.Blocks {
		bc.addBlockTransactionsToPool(*block)
	}
	// restore canonical chain to head
	bc.restoreCanonicalChain(*headCanonicalChain)

	sideChain := bc.getSideChains()[sideChainIndex]
	// add sidechain blocks to blockchain
	for _, newBlock := range sideChain.Blocks {
		bc.addBlock(newBlock, true)
	}
}

func (bc *BlockChain) checkLongestChain() (sidechain, parentBlock int) {
	sideChains := bc.getSideChains()
	canonicalChain := bc.getCanonicalChain()
	for sideChainIndex := 0; sideChainIndex < len(sideChains); sideChainIndex++ {
		sideChain := sideChains[sideChainIndex]
		chainHead := sideChain.Blocks[0]
		for blockIndex := 0; blockIndex < canonicalChain.size(); blockIndex++ {
			block := canonicalChain.Blocks[blockIndex]
			if block.IsNextBlock(chainHead) {
				logger.Logf("Parent of SideChain found in canonicalChain")
				// if a block in the canonical chain is the parent
				// of a head of a chain, lets check the sizes
				headChainSize := blockIndex + 1
				subChainSize := canonicalChain.size() - headChainSize
				// if sidechain is longer, there is a fork,
				// return the index of the restoring sidechein
				// and the heah in the blockchain
				if sideChain.size() > subChainSize {
					logger.Logf("SideChain of size %v found bigger than canonical %v", sideChain.size(), subChainSize)
					sidechain = sideChainIndex
					parentBlock = blockIndex
					return
				}
				logger.LogForkShort(block.String())
			}

		}
	}
	return -1, -1
}

// Stop func
func (bc *BlockChain) Stop() {
	bc.setMining(false)
	if bc.isActive() {
		logger.Logf("Stopping BlockChain handler")
		bc.setActive(false)
		close(bc.quitChannel)
	} else {
		logger.Logf("BlockChain already stopped....")
	}
}

func (bc *BlockChain) GetBalanceOfHash(hash utils.HashValue) int {
	cc := bc.getCanonicalChain()
	var balance int
	for _, bl := range cc.Blocks {
		for _, tx := range bl.Transactions {
			if tx.Input.Equals(hash.String()) {
				balance = balance - int(tx.Amount)
			}
			if tx.Output.Equals(hash.String()) {
				balance = balance + int(tx.Amount)
			}
		}
	}
	return balance
}
