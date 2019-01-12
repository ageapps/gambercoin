package node

import (
	"fmt"

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
	logger.Log(fmt.Sprint("Sending STATUS route: ", nodeName != ""))
	packet := &data.GossipPacket{Status: message}
	node.peerConection.SendPacketToPeer(destination, packet)
}

func (node *Node) sendRumrorMessage(destinationAdress, origin string, id uint32) {
	if message := node.rumorStack.GetRumorMessage(origin, id); message != nil {
		packet := &data.GossipPacket{Rumor: message}
		logger.Log(fmt.Sprintf("Sending RUMOR ID:%v", message.ID))
		node.peerConection.SendPacketToPeer(destinationAdress, packet)
	} else {
		logger.Log("Message to send not found")
	}
}

func (node *Node) sendRouteRumorMessage(destinationAdress string) {
	latestMsgID := node.rumorCounter.GetValue() + 1
	routeRumorMessage := monguer.NewRumorMessage(node.Name, latestMsgID, "")
	packet := &data.GossipPacket{Rumor: routeRumorMessage}
	logger.Log(fmt.Sprintf("Sending ROUTE RUMOR ID:%v", latestMsgID))
	node.peerConection.SendPacketToPeer(destinationAdress, packet)
}

func (node *Node) broadcastRouteRumorMessage(destinationAdress string) {
	latestMsgID := node.rumorCounter.GetValue() + 1
	routeRumorMessage := monguer.NewRumorMessage(node.Name, latestMsgID, "")
	packet := &data.GossipPacket{Rumor: routeRumorMessage}
	logger.Log(fmt.Sprintf("Sending ROUTE RUMOR ID:%v", latestMsgID))
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
