package blockchain

// Chain struct
type Chain struct {
	Blocks []*Block
}

// NewEmptyChain func
func NewEmptyChain() Chain {
	return Chain{Blocks: []*Block{}}
}

// appendBlock to chain
func (chain *Chain) appendBlock(block *Block) {
	chain.Blocks = append(chain.Blocks, block)
}

// size get chain length
func (chain *Chain) size() int {
	return len(chain.Blocks)
}

// isNextBlockInChain check if block is the next in the chain
func (chain *Chain) isNextBlockInChain(newBlock *Block) bool {
	if len(chain.Blocks) <= 0 {
		return true
	}
	lastBlock := chain.Blocks[len(chain.Blocks)-1]
	return lastBlock.IsNextBlock(newBlock)
}

// getSubchain giben start and end index
func (chain *Chain) getSubchain(start, end int) *Chain {
	return &Chain{Blocks: chain.Blocks[start:end]}
}
