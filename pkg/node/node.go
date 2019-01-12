package node

import (
	"fmt"
	"log"
	"sync"

	"github.com/ageapps/gambercoin/pkg/blockchain"
	"github.com/ageapps/gambercoin/pkg/connection"
	"github.com/ageapps/gambercoin/pkg/data"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/monguer"
	"github.com/ageapps/gambercoin/pkg/router"
	"github.com/ageapps/gambercoin/pkg/utils"
)

const (
	// ENTROPY_TIMER_PERIOD in seconds
	ENTROPY_TIMER_PERIOD = 1
	ROUTE_TIMER_PERIOD   = 3
)

// Node struct
type Node struct {
	Name            string
	Address         utils.PeerAddress
	peerConection   *connection.ConnectionHandler
	peers           *utils.PeerAddresses
	rumorStack      RumorStack
	privateStack    PrivateStack
	router          *router.Router
	monguerPocesses map[string]*monguer.MongerHandler
	rumorCounter    *utils.Counter // [name] address
	privateCounter  *utils.Counter // [name] address
	mux             sync.Mutex
	usedPeers       map[string]bool
	running         bool
	chainHandler    *blockchain.ChainHandler
}

// NewNode return new instance
func NewNode(addressStr, name string) (*Node, error) {
	address, err := utils.GetPeerAddress(addressStr)
	if err != nil {
		return nil, err
	}
	logger.Logf("Listening to peers in address <%v>", addressStr)

	return &Node{
		Name:            name,
		Address:         address,
		peers:           utils.EmptyAdresses(),
		rumorStack:      RumorStack{Messages: make(map[string][]monguer.RumorMessage)},
		privateStack:    PrivateStack{Messages: make(map[string][]data.PrivateMessage)},
		router:          router.NewRouter(),
		monguerPocesses: make(map[string]*monguer.MongerHandler),
		rumorCounter:    utils.NewCounter(uint32(0)),
		privateCounter:  utils.NewCounter(uint32(0)),
		usedPeers:       make(map[string]bool),
		running:         true,
	}, nil
}

// Start node process
func (node *Node) Start(clientChan chan data.Message) error {
	connection, err := connection.NewConnectionHandler(node.Address.String(), node.Name, true)
	if err != nil {
		return err
	}
	node.peerConection = connection
	go node.listenToClientChannel(clientChan)
	go node.startRouteTimer(ROUTE_TIMER_PERIOD)
	go node.startEntropyTimer(ENTROPY_TIMER_PERIOD)
	// TODO start blockchain process
	return node.listenToPeers()
}

// Stop node process
func (node *Node) Stop() {
	logger.Log("Finishing Node " + node.Name)
	node.setRunning(false)
	for _, process := range node.getMongerProcesses() {
		process.Stop()
	}
	node.peerConection.Close()
	// node = nil
}

// listenToClientChannel function
// start to listen for client messages
// in input channel
func (node *Node) listenToClientChannel(clientChan chan data.Message) {
	if clientChan == nil {
		logger.Log("No client input channel created")
		return
	}
	for msg := range clientChan {
		if msg.Destination == "" {
			log.Fatal("message received is not valid")
		}
		node.handleClientMessage(&msg)
	}
}

// listenToPeers function
// Start listening to Packets from peers
func (node *Node) listenToPeers() error {
	if node.peerConection == nil {
		return fmt.Errorf("node not connected to peers")
	}
	for pkt := range node.peerConection.MessageQueue {
		if pkt.Address == "" {
			return fmt.Errorf("message received is not valid")
		}
		node.handlePeerPacket(&pkt.Packet, pkt.Address)
	}
	return nil
}

// handleClientMessage handles client messages
func (node *Node) handleClientMessage(msg *data.Message) {

	logger.Logf("Message received from client \nprivate: %v", msg.IsDirectMessage())

	switch {
	case msg.Broadcast:
		logger.LogClient((*msg).Text)

		newMsg := data.NewSimpleMessage(node.Name, msg.Text, node.Address.String())
		node.peerConection.BroadcastPacket(node.peers, &data.GossipPacket{Simple: newMsg}, node.Address.String())

	case msg.IsDirectMessage():
		node.handleClientDirectMessage(msg)

	default:
		logger.LogClient((*msg).Text)
		// Reset used peers for timers
		go node.resetUsedPeers()
		id := node.rumorCounter.Increment()
		rumorMessage := monguer.NewRumorMessage(node.Name, id, msg.Text)
		node.rumorStack.AddMessage(*rumorMessage)
		node.mongerMessage(rumorMessage, "", false)
	}
}

func (node *Node) handleClientDirectMessage(msg *data.Message) {
	// Message has request hash
	// Message is a private message
	logger.LogClient((*msg).Text)
	// Message is private
	id := node.privateCounter.Increment()
	privateMessage := data.NewPrivateMessage(node.Name, id, msg.Destination, msg.Text, uint32(10))
	node.privateStack.AddMessage(*privateMessage)
	node.sendPrivateMessage(privateMessage)
}

func (node *Node) handlePeerPacket(packet *data.GossipPacket, originAddress string) {
	if originAddress != node.Address.String() {
		err := node.GetPeers().Set(originAddress)
		if err != nil {
			log.Fatal(err)
		}
	}

	packetType := packet.GetPacketType()
	logger.Log("Received packet peer: " + packetType)
	switch packetType {
	case data.PACKET_STATUS:
		node.handleStatusMessage(packet.Status, originAddress)
	case data.PACKET_RUMOR:
		node.handleRumorMessage(packet.Rumor, originAddress)
	case data.PACKET_PRIVATE:
		node.handlePeerPrivateMessage(packet.Private, originAddress)
	case data.PACKET_TX:
		node.handleTxMessage(packet.TxMessage, originAddress)
	case data.PACKET_BLOCK:
		node.handleBlockMessage(packet.BlockMessage, originAddress)
	case data.PACKET_SIMPLE:
		msg := *packet.Simple
		logger.LogSimple(msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
		logger.LogPeers(node.peers.String())
		node.handleSimpleMessage(packet.Simple, originAddress)
	default:
		logger.Log("Message not recognized")
		// log.Fatal(errors.New("Message not recognized"))
	}
}

func (node *Node) mongerMessage(msg *monguer.RumorMessage, originPeer string, routerMonguering bool) {
	node.mux.Lock()
	name := fmt.Sprint(len(node.monguerPocesses), "/", routerMonguering)
	monguerProcess := monguer.NewMongerHandler(originPeer, name, routerMonguering, msg, node.peers)
	node.mux.Unlock()

	node.registerProcess(monguerProcess, PROCESS_MONGUER)
	monguerProcess.Start(func() {
		node.unregisterProcess(monguerProcess.Name, PROCESS_MONGUER)
	})
}