package client

import (
	"github.com/ageapps/gambercoin/pkg/blockchain"
	"github.com/ageapps/gambercoin/pkg/utils"
)

// Message to send
type Message struct {
	Text        string
	Destination string
	Broadcast   bool
	Transaction *ClientTx
}

// ClientTx to send
type ClientTx struct {
	In     string
	Out    string
	Amount int
}

func (tx *ClientTx) GetTransaction() blockchain.Transaction {
	in, _ := utils.GetHash(tx.In)
	out, _ := utils.GetHash(tx.Out)
	return blockchain.NewTransaction(in, out, tx.Amount)
}

// IsDirectMessage check if is private message
func (msg *Message) IsDirectMessage() bool {
	return msg.Destination != ""
}

// IsTx check if is Tx message
func (msg *Message) IsTx() bool {
	return msg.Transaction != nil
}
