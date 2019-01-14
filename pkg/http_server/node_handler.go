package http_server

import (
	"log"
	"strings"
	"sync"

	"github.com/ageapps/gambercoin/pkg/stack"

	"github.com/ageapps/gambercoin/pkg/client"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/node"
	"github.com/ageapps/gambercoin/pkg/router"
	"github.com/ageapps/gambercoin/pkg/utils"
)

func NewNodePool() *NodePool {
	return &NodePool{
		nodes:           make(map[string]*node.Node),
		messageChannels: make(map[string]*chan client.Message),
	}
}

// NodePool struct cointaining nodes
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
func (pool *NodePool) deleteNode(name string) {
	pool.mux.Lock()
	delete(pool.nodes, name)
	pool.mux.Unlock()
}
func (pool *NodePool) getNode(name string) (foundNode *node.Node, found bool) {
	pool.mux.Lock()
	defer pool.mux.Unlock()
	foundNode, found = pool.nodes[name]
	return
}
func (pool *NodePool) getMsgChannel(name string) (foundChannel chan client.Message, found bool) {
	pool.mux.Lock()
	defer pool.mux.Unlock()
	foundChannelPtr, found := pool.messageChannels[name]
	return *foundChannelPtr, found
}
func (pool *NodePool) findNode(name, address string) (*node.Node, bool) {
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
	nodePool = NewNodePool()
	// DebugLevel default level
	DebugLevel = logger.Info
)

// StatusResponse struct
type StatusResponse struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

func startNode(name, address string, peers *utils.PeerAddresses) string {
	shortName := strings.Split(name, "-")[0]

	logger.CreateLogger(shortName, strings.Split(address, ":")[1], DebugLevel)
	targetNode, found := nodePool.findNode(name, address)

	if !found {
		newNode, err := node.NewNode(address, name)
		if err != nil {
			logger.Logw("Error creating new Node, %v ", err)
			return ""
		}
		targetNode = newNode
		if peers != nil && len(peers.GetAdresses()) > 0 {
			go targetNode.AddPeers(peers)
		}
		messageChannel := make(chan client.Message)
		go func(channel chan client.Message) {
			if err := targetNode.Start(channel); err != nil {
				log.Fatal(err)
			}
		}(messageChannel)

		// go func() {
		// 	targetNode.StartBlockChain()
		// }()
		nodePool.addNode(targetNode)
		nodePool.addMsgChannel(targetNode.Name, &messageChannel)
	}
	return targetNode.Name
}

func getNodeRoutes(name string) *router.RoutingTable {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return nil
	}

	return targetNode.GetRoutes()
}

func getNodeMessages(name string) *[]stack.GenericMessage {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return nil
	}
	return targetNode.GetLatestMessages()
}

func getNodePrivateMessages(name string) *map[string][]stack.GenericMessage {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return nil
	}
	return targetNode.GetPrivateMessages()
}

func getNodePeers(name string) *[]string {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return nil
	}
	return targetNode.GetPeerArray()
}
func geetHashBalance(name string, hash utils.HashValue) int {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return -1000000
	}
	return targetNode.GetBalanceOfHash(hash)
}

func getStatusResponse(name string) *StatusResponse {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return nil
	}
	return &StatusResponse{
		Name:    targetNode.Name,
		Address: targetNode.Address.String(),
	}
}

func deleteNode(name string) {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return
	}
	go targetNode.Stop()
	nodePool.deleteNode(targetNode.Name)
}

func addPeer(name, peer string) bool {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return false
	}
	targetNode.AddAndNotifyPeer(peer)
	return true
}

func sendMessage(name, msg string) bool {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return false
	}
	newMsg := &client.Message{
		Text: msg,
	}
	channel, _ := nodePool.getMsgChannel(targetNode.Name)
	channel <- *newMsg
	return true
}

func sendPrivateMessage(name, destination, msg string) bool {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return false
	}
	newMsg := &client.Message{
		Text:        msg,
		Destination: destination,
	}
	channel, _ := nodePool.getMsgChannel(targetNode.Name)
	channel <- *newMsg
	return true
}

func sendTransaction(name, in, out string, amount int) bool {
	targetNode, found := nodePool.getNode(name)
	if !found {
		return false
	}
	newMsg := &client.Message{
		Transaction: &client.ClientTx{in, out, amount},
	}
	channel, _ := nodePool.getMsgChannel(targetNode.Name)
	channel <- *newMsg
	return true
}
