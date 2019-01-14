package node

import (
	"log"

	"github.com/ageapps/gambercoin/pkg/logger"

	"github.com/ageapps/gambercoin/pkg/router"

	"github.com/ageapps/gambercoin/pkg/stack"

	"github.com/ageapps/gambercoin/pkg/utils"
)

// AddPeers peers
func (node *Node) AddPeers(newPeers *utils.PeerAddresses) {
	if newPeers != nil && len(newPeers.GetAdresses()) > 0 {
		node.mux.Lock()
		node.peers.AppendPeers(newPeers)
		logger.LogPeers(node.peers.String())
		node.mux.Unlock()
	}
}

// AddAndNotifyPeer func
func (node *Node) AddAndNotifyPeer(newPeer string) {
	err := node.peers.Set(newPeer)
	if err != nil {
		log.Fatal(err)
	}
	node.sendStatusMessage(newPeer, "")
}

// GetLatestMessages returns last rumor messages
func (node *Node) GetLatestMessages() *[]stack.GenericMessage {
	return node.rumorStack.GetLatestMessages()
}

// GetPrivateMessages returns last private messages
func (node *Node) GetPrivateMessages() *map[string][]stack.GenericMessage {
	return node.privateStack.GetStack()
}

// GetPeerArray returns an array of address strings
func (node *Node) GetPeerArray() *[]string {
	var peersArr = []string{}
	for _, peer := range node.GetPeers().GetAdresses() {
		peersArr = append(peersArr, peer.String())
	}
	return &peersArr
}

// GetPeers returns current peers
func (node *Node) GetPeers() *utils.PeerAddresses {
	node.mux.Lock()
	defer node.mux.Unlock()
	return node.peers
}

// IsRunning node
func (node *Node) IsRunning() bool {
	node.mux.Lock()
	defer node.mux.Unlock()
	return node.running
}

// Stop node
func (node *Node) setRunning(running bool) {
	node.mux.Lock()
	node.running = running
	node.mux.Unlock()
}

// GetRoutes returns the routing table
func (node *Node) GetRoutes() *router.RoutingTable {
	node.mux.Lock()
	defer node.mux.Unlock()

	return node.router.GetTable()
}

func (node *Node) resetUsedPeers() {
	node.mux.Lock()
	node.usedPeers = make(map[string]bool)
	node.mux.Unlock()
}

// GetUsedPeers funct
func (node *Node) GetUsedPeers() map[string]bool {
	node.mux.Lock()
	defer node.mux.Unlock()
	return node.usedPeers
}

// ReceivedRouteAck funct
func (node *Node) ReceivedRouteAck() bool {
	node.mux.Lock()
	defer node.mux.Unlock()
	return node.receivedRoute
}

// SetReceivedRoute funct
func (node *Node) SetReceivedRoute(received bool) {
	node.mux.Lock()
	defer node.mux.Unlock()
	node.receivedRoute = received
}

// GetBalanceOfHash funct
func (node *Node) GetBalanceOfHash(hash utils.HashValue) int {
	node.mux.Lock()
	defer node.mux.Unlock()
	return node.blockchain.GetBalanceOfHash(hash)
}
