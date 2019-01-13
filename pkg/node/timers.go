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

// startRouteTimer function
// this timer sends preriodically a route message to a random peer
func (node *Node) startRouteTimer(rtimer int) {
	// logger.Log("Starting Route timer")
	usedPeers := make(map[string]bool)
	for node.IsRunning() {
		if newpeer := node.GetPeers().GetRandomPeer(usedPeers); newpeer != nil {
			logger.Logv("Route Timer - MESSAGE")
			node.sendRouteRumorMessage(newpeer.String())
		}
		time.Sleep(time.Duration(rtimer) * time.Second)
	}
}
