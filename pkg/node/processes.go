package node

import (
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/monguer"
)

// ProcessType string
type ProcessType string

const (
	// PROCESS_MONGUER var
	PROCESS_MONGUER ProcessType = "PROCESS_MONGUER"
)

func (node *Node) registerMonguerProcess(process *monguer.MongerHandler) {
	node.mux.Lock()
	node.monguerPocesses[process.Name] = process
	node.mux.Unlock()
	logger.Logf("%v - Registering %v - %v", node.Name, PROCESS_MONGUER, process.Name)
}

func (node *Node) unregisterProcess(name string) {
	found := true
	_, found = node.monguerPocesses[name]
	node.mux.Lock()
	node.monguerPocesses[name] = nil
	delete(node.monguerPocesses, name)
	node.mux.Unlock()
	logger.Logf("%v - Unregistering %v - %v found:%v", node.Name, PROCESS_MONGUER, name, found)
}

func (node *Node) processExists(name string) bool {
	node.mux.Lock()
	_, exists := node.monguerPocesses[name]
	node.mux.Unlock()
	return exists
}

func (node *Node) getMongerProcesses() map[string]*monguer.MongerHandler {
	node.mux.Lock()
	defer node.mux.Unlock()
	return node.monguerPocesses
}

func (node *Node) findMonguerProcess(originAddress string, routeMonguer bool) *monguer.MongerHandler {
	processes := node.getMongerProcesses()
	node.mux.Lock()
	defer node.mux.Unlock()

	for _, process := range processes {
		if process.GetMonguerPeer() == originAddress && process.IsRouteMonguer() == routeMonguer {
			return process
		}
	}
	return nil
}
