package monguer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/signal"
	"github.com/ageapps/gambercoin/pkg/utils"
)

var usedPeers = make(map[string]bool)

// MongerHandler is a handler that will be in
// charge of the monguering process whenever the
// node gets a message from a client:
// name                   name of the Handler
// originalMessage        original message that was being monguered
// currentMessage         message currently being monguered
// currentPeer            client currently being monguered
// active                 monguer handler active state
// currentlySynchronicing bool
// connection             *ConnectionHandler
// peers                  *utils.PeerAddresses
// mux                    sync.Mutex
// timer                  *time.Timer
// quitChannel            chan bool
// resetChannel           chan bool
//
type MongerHandler struct {
	Name          string
	SendChannel   chan<- MongerBundle // write-only channel to send messages to node
	SignalChannel chan signal.Signal  // channel to receive messages from node

	originPeer             string
	currentMessage         *RumorMessage
	currentPeer            string
	active                 bool
	routeMonguer           bool
	currentlySynchronicing bool
	peers                  *utils.PeerAddresses
	timer                  *time.Timer
	quitChannel            chan bool
	resetChannel           chan bool
	usedPeers              *map[string]bool
	sync.Mutex
}

// NewMongerHandler function
func NewMongerHandler(originPeer, Name string, isRouter bool, msg *RumorMessage, connectPeers *utils.PeerAddresses) *MongerHandler {
	used := make(map[string]bool)
	if originPeer != "" {
		used[originPeer] = true
	}
	return &MongerHandler{
		originPeer:             originPeer,
		Name:                   Name,
		currentMessage:         msg,
		currentPeer:            "",
		active:                 false,
		routeMonguer:           isRouter,
		currentlySynchronicing: false,
		peers:                  connectPeers,
		timer:                  &time.Timer{},
		quitChannel:            make(chan bool),
		resetChannel:           make(chan bool),
		usedPeers:              &used,
	}
}

// Start monguering process
func (handler *MongerHandler) Start(onStopHandler func()) <-chan MongerBundle {
	sendChannel := make(chan MongerBundle)
	handler.SignalChannel = make(chan signal.Signal)
	handler.SendChannel = sendChannel
	go func() {
		handler.setActive(true)
		handler.monguerWithPeer(false)
		for {
			select {
			case <-handler.resetChannel:
				logger.Logf("Restarting monger handler - %v", handler.Name)
				handler.monguerWithPeer(true)

			case s := <-handler.SignalChannel:
				switch s {
				case signal.Stop:
					handler.stop()
				case signal.Reset:
					handler.reset()
				case signal.Sync:
					handler.setSynking(true)
				}
			case <-handler.timer.C:
				// Flip coin
				if !handler.isSynking() {
					logger.Logf("TIMEOUT, FLIPPING COIN")
					if !keepRumorering() {
						handler.stop()
					} else {
						handler.resetUsedPeers()
						handler.monguerWithPeer(true)
					}
				}
			case <-handler.quitChannel:
				logger.Logf("Finishing monger handler - " + handler.Name)
				if handler.timer.C != nil {
					handler.timer.Stop()
				}
				close(handler.resetChannel)
				onStopHandler()
				return
			}
		}
	}()
	return sendChannel
}

func (handler *MongerHandler) newTimer() *time.Timer {
	// logger.Logf("Launching new timer")
	return time.NewTimer(1 * time.Second)
}

func (handler *MongerHandler) resetUsedPeers() {
	handler.Lock()
	defer handler.Unlock()

	used := make(map[string]bool)
	if handler.originPeer != "" {
		used[handler.originPeer] = true
	}
	handler.usedPeers = &used
}

func (handler *MongerHandler) monguerWithPeer(flipped bool) {
	if peer := handler.getPeers().GetRandomPeer(*handler.usedPeers); peer != nil {
		handler.timer = handler.newTimer()
		handler.setMonguerPeer(peer.String())
		handler.addUsedPeer(peer.String())
		// logger.Logf(fmt.Sprint("Monguering with peer: ", peer.String()))
		if !flipped {
			logger.LogMonguer(peer.String())
		} else {
			logger.LogCoin(peer.String())
		}
		handler.SendChannel <- MongerBundle{handler.getMonguerMessage(), peer.String()}
	} else {
		logger.Logf(fmt.Sprint("No peers to monger with"))
		handler.stop()
	}
}

// Stop handler
func (handler *MongerHandler) stop() {
	handler.setSynking(false)
	if handler.isActive() {
		logger.Logf("Stopping monger handler - %v", handler.Name)
		handler.setActive(false)
		close(handler.quitChannel)
	} else {
		logger.Logf("Monguer process is not active...")
	}
}

// Reset handler
func (handler *MongerHandler) reset() {
	handler.setSynking(false)
	handler.Lock()
	defer handler.Unlock()
	logger.Logf("Restart monger handler")
	go func() {
		handler.resetChannel <- true
	}()
}

//GetMonguerMessage function
func (handler *MongerHandler) getMonguerMessage() *RumorMessage {
	handler.Lock()
	defer handler.Unlock()
	// logger.Logf(fmt.Sprint("Monger message is ", handler.currentMessage))
	return handler.currentMessage
}

// GetMonguerPeer function
func (handler *MongerHandler) GetMonguerPeer() string {
	handler.Lock()
	defer handler.Unlock()
	return handler.currentPeer
}

// IsRouteMonguer gets handler status
func (handler *MongerHandler) IsRouteMonguer() bool {
	handler.Lock()
	defer handler.Unlock()
	return handler.routeMonguer
}

// getPeers function
func (handler *MongerHandler) getPeers() *utils.PeerAddresses {
	handler.Lock()
	defer handler.Unlock()
	return handler.peers
}

func (handler *MongerHandler) setMonguerPeer(peer string) {
	handler.Lock()
	handler.currentPeer = peer
	handler.Unlock()
}

// IsActive gets handler status
func (handler *MongerHandler) isActive() bool {
	handler.Lock()
	defer handler.Unlock()
	return handler.active
}

func (handler *MongerHandler) isSynking() bool {
	handler.Lock()
	defer handler.Unlock()
	return handler.currentlySynchronicing
}

// SetSynking state
func (handler *MongerHandler) setSynking(value bool) {
	handler.Lock()
	defer handler.Unlock()
	handler.timer.Stop()
	handler.currentlySynchronicing = value
}
func (handler *MongerHandler) setActive(value bool) {
	handler.Lock()
	handler.active = value
	handler.Unlock()
}
func (handler *MongerHandler) addUsedPeer(peer string) {
	handler.Lock()
	(*handler.usedPeers)[peer] = true
	handler.Unlock()
}

func (handler *MongerHandler) logMonguer(msg string) {
	logger.Logf(fmt.Sprintf("[MONGER-%v]%v", handler.Name, msg))
}

func keepRumorering() bool {
	// flipCoin
	coin := rand.Int() % 2
	return coin != 0
}
