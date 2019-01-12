package router

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/utils"
)

// Router struct
type Router struct {
	table RoutingTable
	mux   sync.Mutex
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
	router.mux.Lock()
	defer router.mux.Unlock()
	return &router.table
}

// SetEntry adds entry if there's none for the origin address or there's a new one
func (router *Router) SetEntry(origin, address string) bool {
	router.mux.Lock()
	defer router.mux.Unlock()
	isNew := false
	oldValue, ok := router.table[origin]

	if !ok || oldValue.String() != address {
		isNew = true
		newEntry := utils.PeerAddress{}
		err := newEntry.Set(address)
		if err != nil {
			logger.Log("Error updating router entry")
			return false
		}
		router.addEntry(origin, &newEntry)
	}
	return isNew
}

func (router *Router) addEntry(origin string, entry *utils.PeerAddress) {
	logger.Log(fmt.Sprintf("Route entry appended - Origin:%v", origin))
	router.table[origin] = entry
	logger.LogDSDV(origin, entry.String())
}

// GetAddress returns de addess gibben an identifier
func (router *Router) GetAddress(origin string) (entry *utils.PeerAddress, found bool) {
	router.mux.Lock()
	defer router.mux.Unlock()

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
	router.mux.Lock()
	defer router.mux.Unlock()

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
