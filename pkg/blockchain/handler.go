package blockchain

import (
	"sync"
	"time"

	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/utils"
)

// ChainHandler is a handler that will be in
// charge of requesting from other peers
// FileName            string
// MetaHash            HashValue
// stopped              bool
// connection          *ConnectionHandler
// router              *router.Router
// mux                 sync.Mutex
// quitChannel         chan bool
// resetChannel        chan bool
//
type ChainHandler struct {
	blockchain     *BlockChain
	gossiperAddres string
	stopped        bool
	Peers          *utils.PeerAddresses
	mux            sync.Mutex
	timer          *time.Timer
	quitChannel    chan bool
	BundleChannel  chan *TransactionBundle
	BlockChannel   chan *BlockBundle
}

// NewChainHandler function
func NewChainHandler(address string, peers *utils.PeerAddresses) *ChainHandler {
	return &ChainHandler{
		blockchain:    NewBlockChain(),
		stopped:       false,
		Peers:         peers,
		timer:         &time.Timer{},
		quitChannel:   make(chan bool),
		BundleChannel: make(chan *TransactionBundle),
		BlockChannel:  make(chan *BlockBundle),
	}
}

func (handler *ChainHandler) resetTimer() {
	//logger.Log("Launching new timer")
	if handler.getTimer().C != nil {
		handler.getTimer().Stop()
	}
	handler.mux.Lock()
	handler.timer = time.NewTimer(5 * time.Second)
	handler.mux.Unlock()
}
func (handler *ChainHandler) getTimer() *time.Timer {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	return handler.timer
}

// Start handler
func (handler *ChainHandler) Start(onStopHandler func()) {
	go handler.resetTimer()

	go func() {
		for {
			select {
			case bundle := <-handler.BundleChannel:
				transactionMessage := *bundle.TxMessage
				transaction := &transactionMessage.Tx

				if !handler.blockchain.IsTransactionSaved(transaction) {
					transactionMessage.HopLimit--
					handler.addTransaction(transaction)
					if transactionMessage.HopLimit > 0 {
						// handler.publishTX(file, transaction.HopLimit, bundle.Origin)
					}
				} else {
					logger.Log("Transaction for already indexed" + transaction.String())
				}
			case minedBlock := <-handler.blockchain.MinedBlocks:
				timer2 := time.NewTimer(time.Duration(2 * handler.blockchain.getBlockTime()))
				go func() {
					<-timer2.C
					handler.publishBlock(minedBlock, uint32(20), handler.gossiperAddres)
				}()

			case blockBundle := <-handler.BlockChannel:
				blockMsg := blockBundle.BlockMessage
				block := blockMsg.Block
				logger.Logf("Received block - %v", block.String())
				blockMsg.HopLimit--
				if handler.blockchain.CanAddBlock(&block) {
					handler.addBlock(&block)
					handler.indexTransactionsInBlock(&blockMsg.Block)
					if blockMsg.HopLimit > 0 {
						handler.publishBlock(&blockMsg.Block, blockMsg.HopLimit, blockBundle.Origin)
					}
				}
			case <-handler.quitChannel:
				logger.Log("Finishing Blockchain handler")
				if handler.timer.C != nil {
					handler.timer.Stop()
				}
				handler.blockchain.Stop()
				close(handler.BundleChannel)
				close(handler.BlockChannel)
				close(handler.quitChannel)
				onStopHandler()
				return
			}
		}
	}()
}

//StartBlockchain process
func (handler *ChainHandler) StartBlockchain() {
	handler.blockchain.Start(func() {
		logger.Log("Stopped Blockchain succesfully")
	})
}

func (handler *ChainHandler) indexTransactionsInBlock(block *Block) {
	// for _, tx := range block.Transactions {
	// 	handler.GetStore().IndexFile(tx.File)
	// }
}

func (handler *ChainHandler) addTransaction(tx *Transaction) {
	handler.mux.Lock()
	handler.blockchain.TxChannel <- tx
	handler.mux.Unlock()
}
func (handler *ChainHandler) addBlock(bl *Block) {
	handler.mux.Lock()
	handler.blockchain.BlockChannel <- bl
	handler.mux.Unlock()
}

// Stop func
func (handler *ChainHandler) Stop() {
	logger.Log("Stopping handler")
	if !handler.stopped {
		handler.stopped = true
		return
	}
	logger.Log("Data Handler already stopped....")
}

func (handler *ChainHandler) publishTX(hops uint32, origin string) {
	// fmt.Println(file)
	// msg := NewTXPublish(file, hops)
	// fmt.Println(msg.File)
	// packet := &GossipPacket{TxPublish: msg}
}

func (handler *ChainHandler) publishBlock(block *Block, hops uint32, origin string) {
	// logger.Logf("%v", handler.Peers.GetAdresses())
}
