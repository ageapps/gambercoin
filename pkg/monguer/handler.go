package monguer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ageapps/gambercoin/pkg/logger"
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
	Name                   string
	SendChannel            chan MongerBundle
	originPeer             string
	currentMessage         *RumorMessage
	currentPeer            string
	active                 bool
	routeMonguer           bool
	currentlySynchronicing bool
	peers                  *utils.PeerAddresses
	mux                    sync.Mutex
	timer                  *time.Timer
	quitChannel            chan bool
	resetChannel           chan bool
	usedPeers              *map[string]bool
}

// NewMongerHandler function
func NewMongerHandler(originPeer, nameStr string, isRouter bool, msg *RumorMessage, connectPeers *utils.PeerAddresses) *MongerHandler {
	used := make(map[string]bool)
	if originPeer != "" {
		used[originPeer] = true
	}
	return &MongerHandler{
		originPeer:             originPeer,
		Name:                   nameStr,
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
func (handler *MongerHandler) Start(onStopHandler func()) {
	go func() {
		handler.setActive(true)
		handler.monguerWithPeer(false)
		for {
			select {
			case <-handler.resetChannel:
				logger.Logf("Restarting monger handler - %v", handler.Name)
				handler.monguerWithPeer(true)
			case <-handler.timer.C:
				// Flip coin
				if !handler.isSynking() {
					logger.Logf("TIMEOUT, FLIPPING COIN")
					if !keepRumorering() {
						handler.Stop()
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
}

func (handler *MongerHandler) newTimer() *time.Timer {
	// logger.Logf("Launching new timer")
	return time.NewTimer(1 * time.Second)
}

func (handler *MongerHandler) resetUsedPeers() {
	handler.mux.Lock()
	defer handler.mux.Unlock()

	used := make(map[string]bool)
	if handler.originPeer != "" {
		used[handler.originPeer] = true
	}
	handler.usedPeers = &used
}

func (handler *MongerHandler) monguerWithPeer(flipped bool) {
	if peer := handler.GetPeers().GetRandomPeer(*handler.usedPeers); peer != nil {
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
		handler.Stop()
	}
}

// Stop handler
func (handler *MongerHandler) Stop() {
	if handler.IsActive() {
		logger.Logf("Stopping monger handler - %v", handler.Name)
		handler.setActive(false)
		close(handler.quitChannel)
	} else {
		logger.Logf("Monguer process is not active...")
	}
}

// Reset handler
func (handler *MongerHandler) Reset() {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	logger.Logf("Restart monger handler")
	go func() {
		handler.resetChannel <- true
	}()
}

//SetMonguerMessage function
func (handler *MongerHandler) SetMonguerMessage(msg *RumorMessage) {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	handler.currentMessage = msg
}

//GetMonguerMessage function
func (handler *MongerHandler) getMonguerMessage() *RumorMessage {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	// logger.Logf(fmt.Sprint("Monger message is ", handler.currentMessage))
	return handler.currentMessage
}

// GetMonguerPeer function
func (handler *MongerHandler) GetMonguerPeer() string {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	return handler.currentPeer
}

// GetPeers function
func (handler *MongerHandler) GetPeers() *utils.PeerAddresses {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	return handler.peers
}

func (handler *MongerHandler) setMonguerPeer(peer string) {
	handler.mux.Lock()
	handler.currentPeer = peer
	handler.mux.Unlock()
}

// IsActive gets handler status
func (handler *MongerHandler) IsActive() bool {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	return handler.active
}

// IsRouteMonguer gets handler status
func (handler *MongerHandler) IsRouteMonguer() bool {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	return handler.routeMonguer
}

func (handler *MongerHandler) isSynking() bool {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	return handler.currentlySynchronicing
}

// SetSynking state
func (handler *MongerHandler) SetSynking(value bool) {
	handler.mux.Lock()
	defer handler.mux.Unlock()
	handler.timer.Stop()
	handler.currentlySynchronicing = value
}
func (handler *MongerHandler) setActive(value bool) {
	handler.mux.Lock()
	handler.active = value
	handler.mux.Unlock()
}
func (handler *MongerHandler) addUsedPeer(peer string) {
	handler.mux.Lock()
	(*handler.usedPeers)[peer] = true
	handler.mux.Unlock()
}

func (handler *MongerHandler) logMonguer(msg string) {
	logger.Logf(fmt.Sprintf("[MONGER-%v]%v", handler.Name, msg))
}

func keepRumorering() bool {
	// flipCoin
	coin := rand.Int() % 2
	return coin != 0
}
