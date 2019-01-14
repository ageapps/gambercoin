package node

import (
	"time"

	"github.com/ageapps/gambercoin/pkg/logger"
)

// startEntropyTimer function
// This timer sends a status message randomly
// to a known peer that wasn´t notified already
func (node *Node) startEntropyTimer(etimer int) {
	// logger.Log("Starting Entropy timer")
	for node.IsRunning() {
		usedPeers := node.GetUsedPeers()

		if len(usedPeers) >= len(node.GetPeers().GetAdresses()) {
			// logger.Log("Entropy Timer - All peers where notified")
		} else if newpeer := node.GetPeers().GetRandomPeer(usedPeers); newpeer != nil {
			logger.Logv("Entropy Timer - MESSAGE to %v", newpeer.String())
			node.mux.Lock()
			node.usedPeers[newpeer.String()] = true
			node.mux.Unlock()
			node.sendStatusMessage(newpeer.String(), "")
		}
		time.Sleep(time.Duration(etimer) * time.Second)
	}
}

func newTimer(timer int) *time.Timer {
	// logger.Logf("Launching new timer")
	return time.NewTimer(time.Duration(timer) * time.Second)
}

// startRouteTimer function
// this timer sends preriodically a route message to a random peer
func (node *Node) startRouteTimer(rtimer int) {
	// logger.Log("Starting Route timer")
	usedPeers := make(map[string]bool)
	currentPeer := node.sendRandomRouteMessage(usedPeers)
	retryCunt := 0

	for node.IsRunning() {

		if node.ReceivedRouteAck() {
			currentPeer = node.sendRandomRouteMessage(usedPeers)
			retryCunt = 0
			node.SetReceivedRoute(false)
		} else if currentPeer != "" {
			logger.Logi("ROUTE TIMEOUT")
			retryCunt++
			node.sendRouteRumorMessage(currentPeer)
		} else {
			currentPeer = node.sendRandomRouteMessage(usedPeers)
		}

		if retryCunt >= MAX_RETRYS {
			node.GetPeers().RemovePeer(currentPeer)
			logger.Logi("Removed PEER %v", currentPeer)
			currentPeer = node.sendRandomRouteMessage(usedPeers)
			retryCunt = 0
		}
		time.Sleep(time.Duration(rtimer) * time.Second)
	}
}

func (node *Node) sendRandomRouteMessage(usedPeers map[string]bool) string {
	if newpeer := node.GetPeers().GetRandomPeer(usedPeers); newpeer != nil {
		logger.Logv("Route Timer - MESSAGE")
		node.sendRouteRumorMessage(newpeer.String())
		return newpeer.String()
	}
	return ""
}
