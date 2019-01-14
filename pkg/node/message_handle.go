package node

import (
	"fmt"

	"github.com/ageapps/gambercoin/pkg/blockchain"
	"github.com/ageapps/gambercoin/pkg/signal"
	"github.com/ageapps/gambercoin/pkg/stack"
	"github.com/ageapps/gambercoin/pkg/utils"

	"github.com/ageapps/gambercoin/pkg/data"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/monguer"
)

func (node *Node) handleSimpleMessage(msg *data.SimpleMessage, address string) {
	if msg.OriginalName == node.Name {
		logger.Logv("Received own message")
		return
	}
	newMsg := data.NewSimpleMessage(msg.OriginalName, msg.Contents, node.Address.String())
	node.peerConection.BroadcastPacket(node.peers, &data.GossipPacket{Simple: newMsg}, msg.RelayPeerAddr)
}

func (node *Node) handlePeerPrivateMessage(msg *data.PrivateMessage, address string) {
	if msg.Destination == node.Name {
		node.privateStack.AddMessage(*msg)
		logger.LogPrivate((*msg).Origin, fmt.Sprint((*msg).HopLimit), (*msg).Text)
		return
	}
	msg.HopLimit--
	if msg.HopLimit > 0 {
		node.sendPrivateMessage(msg)
	}
}

func (node *Node) handleRumorMessage(msg *monguer.RumorMessage, address string) {
	msgStatus := node.rumorStack.CompareMessage(msg.Origin, msg.ID)
	isRouteRumor := msg.IsRouteRumor()
	routeNode := ""
	if isRouteRumor {
		logger.Logv("Received ROUTE RUMOR")
		routeNode = msg.Origin
	} else {
		logger.LogRumor((*msg).Origin, address, fmt.Sprint((*msg).ID), (*msg).Text)
	}
	logger.LogPeers(node.peers.String())

	node.router.SetEntry(msg.Origin, address)

	if msgStatus == stack.NEW_MESSAGE {

		// If I get own messages that i didn´t
		// have, set internal rumorCounter
		if msg.Origin == node.Name && node.rumorCounter.GetValue() > msg.ID {
			node.rumorCounter.SetValue(msg.ID)
			return
		}

		if !isRouteRumor {
			// Reset used peers for timers
			go node.resetUsedPeers()

			// message is new
			// -> add it to stack
			node.rumorStack.AddMessage(*msg)
		}
		// -> acknowledge message
		node.sendStatusMessage(address, routeNode)
		// -> start monguering message
		node.mongerMessage(msg, address, isRouteRumor)
	} else if !isRouteRumor {
		// message received is not new
		// send my status msg
		node.sendStatusMessage(address, "")
	}
}

func (node *Node) handleStatusMessage(msg *monguer.StatusPacket, address string) {

	isRouteStatus := msg.IsRouteStatus()
	handler := node.findMonguerProcess(address, isRouteStatus)
	logger.Logv("Handler found: <%v>", handler != nil)
	logger.Logv("STATUS received Route: %v", isRouteStatus)

	if isRouteStatus {
		if msg.Route != node.Name {
			node.router.SetEntry(msg.Route, address)
		}
		if handler != nil {
			handler.SignalChannel <- signal.Stop
		}
		return
	}

	if handler != nil {
		handler.SignalChannel <- signal.Sync
	}
	if len(msg.Want) < len(*node.rumorStack.GetStack()) {
		// check messages that i have from other peers that aren´t in the status message
		missingMessage := *node.rumorStack.GetFirstMissingMessage(&msg.Want)
		if missingMessage != nil {
			node.sendRumorMessage(address, missingMessage.GetOrigin(), missingMessage.GetID())
		}
		return
	}
	logStr := ""
	for _, status := range msg.Want {
		logStr += fmt.Sprintf("peer %v nextID %v ", status.Identifier, status.NextID)
	}
	logger.LogStatus(logStr, address)
	logger.LogPeers(node.peers.String())
	inSync := true

	for _, status := range msg.Want {
		messageStatus := node.rumorStack.CompareMessage(status.Identifier, uint32(status.NextID-1))

		switch messageStatus {
		case stack.NEW_MESSAGE:
			// logger.Log("Node needs to update")
			node.sendStatusMessage(address, "")
			break
		case stack.IN_SYNC:
			// logger.Log("Node and Peer have same messages")
		case stack.OLD_MESSAGE:
			// logger.Log("Peer needs to update")
			node.sendRumorMessage(address, status.Identifier, status.NextID)
			break
		}
		inSync = inSync && messageStatus == stack.IN_SYNC
	}
	if inSync {
		logger.LogInSync(address)
		if handler != nil {
			// Flip coin
			logger.Logv("IN SYNC, FLIPPING COIN")
			if !utils.KeepRumorering() {
				handler.SignalChannel <- signal.Stop
			} else {
				handler.SignalChannel <- signal.Reset
			}
		}
	}
}

func (node *Node) handleTxMessage(msg *blockchain.TxMessage, address string) {
	node.chainHandler.BundleChannel <- &blockchain.TransactionBundle{TxMessage: msg, Origin: address}
}
func (node *Node) handleBlockMessage(msg *blockchain.BlockMessage, address string) {
	node.chainHandler.BlockChannel <- &blockchain.BlockBundle{BlockMessage: msg, Origin: address}
}
