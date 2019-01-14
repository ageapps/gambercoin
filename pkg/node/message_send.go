package node

import (
	"github.com/ageapps/gambercoin/pkg/blockchain"
	"github.com/ageapps/gambercoin/pkg/data"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/monguer"
)

func (node *Node) sendStatusMessage(destination, nodeName string) {
	var message *monguer.StatusPacket
	if nodeName != "" {
		message = monguer.NewStatusPacket(nil, nodeName)
	} else {
		message = node.rumorStack.GetStatusMessage()
	}
	logger.Logv("Sending STATUS route: %v", nodeName != "")
	packet := &data.GossipPacket{Status: message}
	node.peerConection.SendPacketToPeer(destination, packet)
}

func (node *Node) sendRumorMessage(destinationAdress, origin string, id uint32) {
	if message := *node.rumorStack.GetMessage(origin, id); message != nil {
		rumor := message.(monguer.RumorMessage)
		packet := &data.GossipPacket{Rumor: &rumor}
		logger.Logv("Sending RUMOR ID:%v", message.GetID())
		node.peerConection.SendPacketToPeer(destinationAdress, packet)
	} else {
		logger.Logw("Message to send not found")
	}
}

func (node *Node) sendRouteRumorMessage(destinationAdress string) {
	routeRumorMessage := monguer.NewRumorMessage(node.Name, uint32(0), "")
	packet := &data.GossipPacket{Rumor: routeRumorMessage}
	logger.Logv("Sending ROUTE RUMOR")
	node.peerConection.SendPacketToPeer(destinationAdress, packet)
}

func (node *Node) sendPrivateMessage(msg *data.PrivateMessage) {
	packet := &data.GossipPacket{Private: msg}
	if destinationAdress, ok := node.router.GetAddress(msg.Destination); ok {
		logger.Logi("Sending PRIVATE Dest:%v", msg.Destination)
		node.peerConection.SendPacketToPeer(destinationAdress.String(), packet)
	} else {
		logger.Logw("INVALID PRIVATE Dest:%v", msg.Destination)
	}
}

func (node *Node) publishTX(tx blockchain.Transaction, hops uint32, origin string) {
	msg := blockchain.NewTxMessage(tx, hops)
	packet := &data.GossipPacket{TxMessage: msg}
	node.peerConection.BroadcastPacket(node.peers, packet, origin)
}

func (node *Node) publishBlock(bl blockchain.Block, hops uint32, origin string) {
	msg := blockchain.NewBlockMessage(bl, hops)
	packet := &data.GossipPacket{BlockMessage: msg}
	node.peerConection.BroadcastPacket(node.peers, packet, origin)
}
