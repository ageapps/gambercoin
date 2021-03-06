package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/ageapps/gambercoin/pkg/client"
	"github.com/ageapps/gambercoin/pkg/utils"
	"github.com/dedis/protobuf"
)

var (
	// Protocol used for messages
	Protocol = "udp"
	// ServerAdress to connect to
	ServerAdress = utils.PeerAddress{IP: net.ParseIP("127.0.0.1")}
)

func sendMessage(msg, dest string) error {
	fmt.Println("Sending <" + msg + "> to address " + ServerAdress.String())
	fmt.Println("Text: " + msg)
	fmt.Println("Destination: " + dest)

	tmsg := &client.Message{
		Text:        msg,
		Destination: dest,
	}
	buf, err1 := protobuf.Encode(tmsg)
	conn, err2 := net.Dial(Protocol, ServerAdress.String())
	defer conn.Close()
	switch {
	case err1 != nil:
		return err1
	case err2 != nil:
		return err2
	}
	conn.Write(buf)
	return nil
}

func main() {

	// Setup flags with this sintax
	// go run . -UIPort=10000 -msg=Hello
	var UIPort = flag.Int("UIPort", 10000, "Port for the UI client")
	var dest = flag.String("Dest", "", "Destination for the private message")
	var msg = flag.String("msg", "", "Message to be sent")

	flag.Parse()
	ServerAdress.Port = int64(*UIPort)

	if e := sendMessage(*msg, *dest); e != nil {
		log.Fatal(e)
	}
}
