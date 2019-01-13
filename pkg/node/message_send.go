package node

import (
	"github.com/ageapps/gambercoin/pkg/data"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/monguer"
)

func (node *Node) sendStatusMessage(destination, nodeName string) {
	var message *monguer.StatusPacket
	if nodeName != "" {
		message = monguer.NewStatusPacket(nil, nodeName)
	} else {
		message = node.rumorStack.getStatusMessage()
	}
	logger.Logi("Sending STATUS route: %v", nodeName != "")
	packet := &data.GossipPacket{Status: message}
	node.peerConection.SendPacketToPeer(destination, packet)
}

func (node *Node) sendRumorMessage(destinationAdress, origin string, id uint32) {
	if message := node.rumorStack.GetRumorMessage(origin, id); message != nil {
		packet := &data.GossipPacket{Rumor: message}
		logger.Logi("Sending RUMOR ID:%v", message.ID)
		node.peerConection.SendPacketToPeer(destinationAdress, packet)
	} else {
		logger.Logi("Message to send not found")
	}
}

func (node *Node) sendRouteRumorMessage(destinationAdress string) {
	latestMsgID := node.rumorCounter.GetValue() + 1
	routeRumorMessage := monguer.NewRumorMessage(node.Name, latestMsgID, "")
	packet := &data.GossipPacket{Rumor: routeRumorMessage}
	logger.Logi("Sending ROUTE RUMOR ID:%v", latestMsgID)
	node.peerConection.SendPacketToPeer(destinationAdress, packet)
}

func (node *Node) broadcastRouteRumorMessage(destinationAdress string) {
	latestMsgID := node.rumorCounter.GetValue() + 1
	routeRumorMessage := monguer.NewRumorMessage(node.Name, latestMsgID, "")
	packet := &data.GossipPacket{Rumor: routeRumorMessage}
	logger.Logi("Sending ROUTE RUMOR ID:%v", latestMsgID)
	node.peerConection.SendPacketToPeer(destinationAdress, packet)
}

func (node *Node) sendPrivateMessage(msg *data.PrivateMessage) {
	packet := &data.GossipPacket{Private: msg}
	if destinationAdress, ok := node.router.GetAddress(msg.Destination); ok {
		logger.Logf("Sending PRIVATE Dest:%v", msg.Destination)
		node.peerConection.SendPacketToPeer(destinationAdress.String(), packet)
	} else {
		logger.Logf("INVALID PRIVATE Dest:%v", msg.Destination)
	}
}
