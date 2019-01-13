package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/ageapps/gambercoin/pkg/client"
	"github.com/ageapps/gambercoin/pkg/connection"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/node"
	"github.com/ageapps/gambercoin/pkg/utils"
)

// Setup flags with this sintax
// node -UIPort=10000 -nodepAddr=127.0.0.1:5000 -name=nodeA -peers=127.0.0.1:5001,127.0.0.1:5002 -simple
// TO TEST
// go run main.go -UIPort=10000 -nodepAddr=127.0.0.1:5000 -name=nodeA -peers=127.0.0.1:5001 -rtimer=3
// go run main.go -UIPort=10001 -nodepAddr=127.0.0.1:5001 -name=nodeB -peers=127.0.0.1:5002 -rtimer=3
// go run main.go -UIPort=10002 -nodepAddr=127.0.0.1:5002 -name=nodeC -peers=127.0.0.1:5000 -rtimer=3

// listen to udp clients sending messages
func listenToUDPClient(address string, outChan chan<- client.Message) *connection.ConnectionHandler {
	udpConnection, err := connection.NewConnectionHandler(address, "client", true)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for pkt := range udpConnection.MessageQueue {
			if pkt.Address == "" {
				log.Fatal("message received is not valid")
			}
			outChan <- pkt.Message
		}
		close(outChan)
	}()
	return udpConnection
}

func main() {

	var peers = utils.EmptyAdresses()
	var nodepAddr = utils.PeerAddress{IP: net.ParseIP("127.0.0.1"), Port: 5000}
	var UIPort = flag.Int("UIPort", 10000, "Define the port to which the client will connect")
	// var rtimer = flag.Int("rtimer", 3, "Route rumors sending period in seconds, 0 to disable")
	var name = flag.String("name", "node", "Define the name of the node")
	flag.Var(peers, "peers", "Define the addreses of the rest of the peers to connect to separeted by a colon")
	flag.Var(&nodepAddr, "nodepAddr", "Define the ip and port to connect and send gossip messages")

	flag.Parse()

	clientAddress := fmt.Sprintf("%v:%v", nodepAddr.String(), UIPort)
	clientChannel := make(chan client.Message)
	clientConn := listenToUDPClient(clientAddress, clientChannel)
	defer clientConn.Close()

	logger.CreateLogger(*name, nodepAddr.String(), logger.Verbose)

	var node, err = node.NewNode(nodepAddr.String(), *name)
	if err != nil {
		log.Fatal(err)
	}

	go node.AddPeers(peers)
	// Start process
	if err := node.Start(clientChannel); err != nil {
		log.Fatal(err)
	}
}
