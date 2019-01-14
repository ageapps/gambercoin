package router

import (
	"math/rand"
	"sync"

	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/utils"
)

// Router struct
type Router struct {
	table RoutingTable
	sync.Mutex
}

// RoutingTable table map
type RoutingTable map[string]*utils.PeerAddress

// NewRouter func
func NewRouter() *Router {
	return &Router{
		table: make(map[string]*utils.PeerAddress),
	}
}

// GetTable returns routing table
func (router *Router) GetTable() *RoutingTable {
	router.Lock()
	defer router.Unlock()
	return &router.table
}

// AddEntry adds entry if there's none for the origin address or there's a new one
func (router *Router) AddEntry(origin, address string) bool {
	isNew := false
	oldValue, ok := router.GetAddress(origin)

	if !ok || oldValue.String() != address {
		isNew = true
		newEntry := utils.PeerAddress{}
		err := newEntry.Set(address)
		if err != nil {
			logger.Logw("Error updating router entry")
			return false
		}
		router.setEntry(origin, newEntry)
	}
	return isNew
}

func (router *Router) setEntry(origin string, entry utils.PeerAddress) {
	logger.Logv("Route entry appended - Origin:%v", origin)

	router.Lock()
	router.table[origin] = &entry
	router.Unlock()

	logger.LogDSDV(origin, entry.String())
}

// GetAddress returns de addess gibben an identifier
func (router *Router) GetAddress(origin string) (entry *utils.PeerAddress, found bool) {
	router.Lock()
	defer router.Unlock()

	value, ok := router.table[origin]
	if !ok || value == nil {
		return nil, false
	}
	return value, true
}

// GetTableSize func
func (router *Router) GetTableSize() int {
	return len(router.table)
}

// GetRandomDestination func
func (router *Router) GetRandomDestination(usedPeers map[string]int) string {
	router.Lock()
	defer router.Unlock()

	var keys []string
	for dest := range router.table {
		keys = append(keys, dest)
	}
	destinationNr := len(keys)
	if len(usedPeers) >= destinationNr {
		return ""
	}
	for {
		randIndex := rand.Int() % destinationNr
		peerDestination := keys[randIndex]
		if _, ok := usedPeers[peerDestination]; !ok {
			return peerDestination
		}
	}
}
