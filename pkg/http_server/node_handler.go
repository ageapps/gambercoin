package http_server

import (
	"log"
	"sync"

	"github.com/ageapps/gambercoin/pkg/client"
	"github.com/ageapps/gambercoin/pkg/data"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/monguer"
	"github.com/ageapps/gambercoin/pkg/node"
	"github.com/ageapps/gambercoin/pkg/router"
	"github.com/ageapps/gambercoin/pkg/utils"
)

func NewNodePool() *NodePool {
	return &NodePool{
		nodes: make(map[string]*node.Node),
	}
}

// GossiperPool struct cointaining nodes
type NodePool struct {
	nodes           map[string]*node.Node
	messageChannels map[string]*chan client.Message
	mux             sync.Mutex
}

func (pool *NodePool) addNode(node *node.Node) {
	pool.mux.Lock()
	pool.nodes[node.Name] = node
	pool.mux.Unlock()
}
func (pool *NodePool) addMsgChannel(name string, msgChan *chan client.Message) {
	pool.mux.Lock()
	pool.messageChannels[name] = msgChan
	pool.mux.Unlock()
}
func (pool *NodePool) deleteGossiper(name string) {
	pool.mux.Lock()
	delete(pool.nodes, name)
	pool.mux.Unlock()
}
func (pool *NodePool) getGossiper(name string) (foundGossiper *node.Node, found bool) {
	pool.mux.Lock()
	defer pool.mux.Unlock()
	foundGossiper, found = pool.nodes[name]
	return
}
func (pool *NodePool) getMsgChannel(name string) (foundChannel chan client.Message, found bool) {
	pool.mux.Lock()
	defer pool.mux.Unlock()
	foundChannelPtr, found := pool.messageChannels[name]
	return *foundChannelPtr, found
}
func (pool *NodePool) findGossiper(name, address string) (*node.Node, bool) {
	pool.mux.Lock()
	defer pool.mux.Unlock()
	for _, node := range pool.nodes {
		if node.Address.String() == address || node.Name == name {
			logger.Logi("Running node found Name:%v Address:%v", name, address)
			return node, true
		}
	}
	return nil, false
}

var (
	gossiperPool = NewNodePool()
	// DebugLevel default level
	DebugLevel = logger.Verbose
)

// StatusResponse struct
type StatusResponse struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

func startGossiper(name, address string, peers *utils.PeerAddresses) string {
	logger.CreateLogger(name, address, DebugLevel)
	targetGossiper, found := gossiperPool.findGossiper(name, address)

	if !found {
		newNode, err := node.NewNode(address, name)
		if err != nil {
			logger.Logw("Error creating new Node, %v ", err)
			return ""
		}
		targetGossiper = newNode
		if peers != nil && len(peers.GetAdresses()) > 0 {
			go targetGossiper.AddPeers(peers)
		}
		messageChannels := make(chan client.Message)
		if err := targetGossiper.Start(messageChannels); err != nil {
			log.Fatal(err)
		}
		// go func() {
		// 	targetGossiper.StartBlockChain()
		// }()
		gossiperPool.addNode(targetGossiper)
		gossiperPool.addMsgChannel(targetGossiper.Name, &messageChannels)
	}
	return targetGossiper.Name
}

func getGossiperRoutes(name string) *router.RoutingTable {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return nil
	}

	return targetGossiper.GetRoutes()
}

func getGossiperMessages(name string) *[]monguer.RumorMessage {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return nil
	}
	return targetGossiper.GetLatestMessages()
}
func getGossiperPrivateMessages(name string) *map[string][]data.PrivateMessage {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return nil
	}
	return targetGossiper.GetPrivateMessages()
}

func getGossiperPeers(name string) *[]string {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return nil
	}
	return targetGossiper.GetPeerArray()
}

func getStatusResponse(name string) *StatusResponse {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return nil
	}
	return &StatusResponse{
		Name:    targetGossiper.Name,
		Address: targetGossiper.Address.String(),
	}
}

func deleteGossiper(name string) {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return
	}
	go targetGossiper.Stop()
	gossiperPool.deleteGossiper(targetGossiper.Name)
}

func addPeer(name, peer string) bool {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return false
	}
	targetGossiper.AddAndNotifyPeer(peer)
	return true
}

func sendMessage(name, msg string) bool {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return false
	}
	newMsg := &client.Message{
		Text: msg,
	}
	channel, _ := gossiperPool.getMsgChannel(targetGossiper.Name)
	channel <- *newMsg
	return true
}

func sendPrivateMessage(name, destination, msg string) bool {
	targetGossiper, found := gossiperPool.getGossiper(name)
	if !found {
		return false
	}
	newMsg := &client.Message{
		Text:        msg,
		Destination: destination,
	}
	channel, _ := gossiperPool.getMsgChannel(targetGossiper.Name)
	channel <- *newMsg
	return true
}
