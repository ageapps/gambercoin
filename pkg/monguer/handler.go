package monguer

import (
	"fmt"
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
type MongerHandler struct {
	Name          string              // name of the Handler
	SendChannel   chan<- MongerBundle // write-only channel to send messages to node
	SignalChannel chan signal.Signal  // channel to receive messages from node

	originPeer     string        // peer that sent rumor message, this avoids monguering with him
	currentMessage *RumorMessage //  message currently being monguered
	currentPeer    string        //   client currently being monguered
	active         bool          //   monguer handler active state
	synchronizing  bool          //   monguer handler synchronizing state
	maxRetrys      int           //   maximal retrys to monguer with a peer
	retrys         int           //   retrys done to monguer with a peer
	timeout        int           //   timeout between messages
	peers          *utils.PeerAddresses
	timer          *time.Timer
	quitChannel    chan bool
	resetChannel   chan bool
	usedPeers      *map[string]bool
	sync.Mutex
}

// NewMongerHandler function
func NewMongerHandler(originPeer, Name string, msg *RumorMessage, connectPeers *utils.PeerAddresses, maxRetrys, timeout int) *MongerHandler {
	used := make(map[string]bool)
	if originPeer != "" {
		used[originPeer] = true
	}
	return &MongerHandler{
		originPeer:     originPeer,
		Name:           Name,
		currentMessage: msg,
		currentPeer:    "",
		active:         false,
		synchronizing:  false,
		maxRetrys:      maxRetrys,
		retrys:         0,
		timeout:        timeout,
		peers:          connectPeers,
		timer:          &time.Timer{},
		quitChannel:    make(chan bool),
		resetChannel:   make(chan bool),
		usedPeers:      &used,
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
				logger.Logv("Restarting monger handler - %v", handler.Name)
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
					logger.Logv("TIMEOUT, FLIPPING COIN")
					if !handler.keepMonguering() {
						handler.stop()
					} else {
						handler.resetUsedPeers()
						handler.monguerWithPeer(true)
					}
				}
			case <-handler.quitChannel:
				logger.Logv("Finishing monger handler - " + handler.Name)
				if handler.timer.C != nil {
					handler.timer.Stop()
				}
				close(handler.resetChannel)
				close(handler.SendChannel)
				onStopHandler()
				return
			}
		}
	}()
	return sendChannel
}

func (handler *MongerHandler) newTimer() *time.Timer {
	// logger.Logf("Launching new timer")
	return time.NewTimer(time.Duration(handler.timeout) * time.Second)
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
		logger.Logi(fmt.Sprint("No peers to monger with"))
		handler.stop()
	}
}

// Stop handler
func (handler *MongerHandler) stop() {
	handler.setSynking(false)
	if handler.isActive() {
		logger.Logv("Stopping monger handler - %v", handler.Name)
		handler.setActive(false)
		close(handler.quitChannel)
	} else {
		logger.Logv("Monguer process is not active...")
	}
}

// Reset handler
func (handler *MongerHandler) reset() {
	handler.setSynking(false)
	handler.Lock()
	defer handler.Unlock()
	logger.Logv("Restart monger handler")
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
	return handler.currentMessage.IsRouteRumor()
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
	return handler.synchronizing
}

// SetSynking state
func (handler *MongerHandler) setSynking(value bool) {
	handler.Lock()
	defer handler.Unlock()
	if handler.timer.C != nil {
		handler.timer.Stop()
	}
	handler.synchronizing = value
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

func (handler *MongerHandler) keepMonguering() bool {
	if handler.retrys >= handler.maxRetrys {
		// Node tryed to many times to contact another peer
		// consider it as failed peer
		handler.peers.RemovePeer(handler.currentPeer)
		logger.Logi("Removed PEER %v", handler.currentPeer)
		logger.LogPeers(handler.peers.String())
		// restart the monguer process
		return false
	}
	handler.retrys++

	return utils.FlipCoin()
}
