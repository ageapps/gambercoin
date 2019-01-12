package blockchain

// TxMessage struct
type TxMessage struct {
	Tx       Transaction
	HopLimit uint32
}

// TransactionBundle struct
type TransactionBundle struct {
	TxMessage *TxMessage
	Origin    string
}

// BlockBundle struct
type BlockBundle struct {
	BlockMessage *BlockMessage
	Origin       string
}

// BlockMessage struct
type BlockMessage struct {
	Block    Block
	HopLimit uint32
}

// NewTxMessage func
func NewTxMessage(tx Transaction, hops uint32) *TxMessage {
	return &TxMessage{tx, hops}
}

// NewBlockMessage func
func NewBlockMessage(block Block, hops uint32) *BlockMessage {
	return &BlockMessage{block, hops}
}
