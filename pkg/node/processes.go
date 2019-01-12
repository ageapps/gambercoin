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
	// PROCESS_DATA var
	PROCESS_DATA ProcessType = "PROCESS_DATA"
	// PROCESS_SEARCH var
	PROCESS_SEARCH ProcessType = "PROCESS_SEARCH"
)

func (node *Node) registerProcess(process interface{}, ptype ProcessType) {
	go func() {

		name := ""
		switch ptype {
		case PROCESS_MONGUER:
			regProcess := process.(*monguer.MongerHandler)
			name = regProcess.Name
			node.mux.Lock()
			node.monguerPocesses[name] = regProcess
			node.mux.Unlock()
			logger.Logf("%v - Registering %v - %v", node.Name, ptype, name)
		}
	}()
}

func (node *Node) unregisterProcess(name string, ptype ProcessType) {
	go func() {
		found := true
		switch ptype {
		case PROCESS_MONGUER:
			_, found = node.monguerPocesses[name]
			node.mux.Lock()
			node.monguerPocesses[name] = nil
			delete(node.monguerPocesses, name)
			node.mux.Unlock()
			logger.Logf("%v - Unregistering %v - %v found:%v", node.Name, ptype, name, found)
		}
	}()
}

func (node *Node) duplicateProcess(name string, ptype ProcessType) bool {
	node.mux.Lock()
	exists := false
	switch ptype {
	case PROCESS_MONGUER:
		_, exists = node.monguerPocesses[name]
	}
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
