package connection

import (
	"errors"
	"net"

	"github.com/ageapps/gambercoin/pkg/client"
	"github.com/ageapps/gambercoin/pkg/data"
	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/utils"
	"github.com/dedis/protobuf"
)

// ConnectionHandler handles
// all connections coding and decoding all packets
type ConnectionHandler struct {
	address      *net.UDPAddr
	conn         *net.UDPConn
	Name         string
	running      bool
	MessageQueue chan data.UDPMessage
}

// NewConnectionHandler function
func NewConnectionHandler(address, name string, listenToPackets bool) (*ConnectionHandler, error) {
	udpAddr, udpConn, err := createConnection(address)
	if err != nil {
		return nil, err
	}

	conHand := &ConnectionHandler{
		address:      udpAddr,
		conn:         udpConn,
		Name:         name,
		running:      false,
		MessageQueue: make(chan data.UDPMessage),
	}

	go conHand.startListening(conHand.MessageQueue, listenToPackets)

	return conHand, nil
}

// Close connection
func (handler *ConnectionHandler) Close() {
	if err := handler.conn.Close(); err != nil {
		logger.Logw("Error closing connection: %v", err)
		// log.Fatal(err1)
	}
	logger.Logi("Closing connection %v", handler.Name)
	handler.Stop()
}

// CreateConnection in address
func createConnection(address string) (*net.UDPAddr, *net.UDPConn, error) {
	logger.Logw("Starting to listen in address: %v", address)
	if udpAddr, err1 := net.ResolveUDPAddr("udp4", address); err1 != nil {
		return nil, nil, err1
	} else if udpConn, err2 := net.ListenUDP("udp4", udpAddr); err2 != nil {
		return nil, nil, err2
	} else {
		return udpAddr, udpConn, nil
	}
}

func (handler *ConnectionHandler) startListening(messages chan data.UDPMessage, listenToPackets bool) {
	handler.running = true
	for handler.running {
		var msg data.UDPMessage
		var err error

		if listenToPackets {
			packet := &data.GossipPacket{}
			address, e := handler.readPacket(packet)

			msg = data.UDPMessage{Packet: *packet, Address: address}
			err = e
		} else {
			message := &client.Message{}
			address, e := handler.readMessage(message)

			msg = data.UDPMessage{Message: *message, Address: address}
			err = e
		}
		if err != nil {
			logger.Logw("Error reading packet")
			break
		}
		go func(running bool) {
			if running {
				messages <- msg
			}
		}(handler.running)
	}
	close(messages)
}

// BroadcastPacket function
func (handler *ConnectionHandler) BroadcastPacket(peers *utils.PeerAddresses, packet *data.GossipPacket, incommingPeer string) {
	logger.Logi("Broadcasting packet %v", packet.GetPacketType())
	for _, peer := range peers.GetAdresses() {
		if incommingPeer == peer.String() {
			continue
		}
		handler.SendPacketToPeer(peer.String(), packet)
	}
}

// SendPacketToPeer sends packet to address
func (handler *ConnectionHandler) SendPacketToPeer(address string, packet *data.GossipPacket) error {
	if handler.conn == nil {
		return errors.New("No connection")
	}
	logger.Logi("Sending packet %v to <%v>", packet.GetPacketType(), address)

	go func() {
		udpaddr, err1 := net.ResolveUDPAddr("udp4", address)
		if err1 != nil {
			logger.Logw("Error Resolving address %v", address)
		}
		packetBytes, err2 := protobuf.Encode(packet)
		if err2 != nil {
			//logger.Logf("Warning Encoding: %v", err2)
		}
		if _, err3 := handler.conn.WriteToUDP(packetBytes, udpaddr); err3 != nil {
			logger.Logw("Error Sending Packet %v", err3)
		}
	}()
	return nil
}

// readPacket reads and decodes packet
func (handler *ConnectionHandler) readPacket(packet *data.GossipPacket) (string, error) {

	if handler.conn == nil {
		return "", errors.New("No connection")
	}
	buffer := make([]byte, 65535)
	_, address, err1 := handler.conn.ReadFromUDP(buffer)
	if err1 != nil {
		logger.Logw("Error Reading packet UDP")
		return "", err1
	}
	err2 := protobuf.Decode(buffer, packet)
	if err2 != nil {
		// logger.Logf("Warning Decoding: %v", err2)
	}
	return address.String(), nil
}

// readMessage reads and decodes message
func (handler *ConnectionHandler) readMessage(msg *client.Message) (string, error) {
	buffer := make([]byte, 1024)
	_, address, err := handler.conn.ReadFromUDP(buffer)
	protobuf.Decode(buffer, msg)
	return address.String(), err
}

// Stop handler
func (handler *ConnectionHandler) Stop() {
	handler.running = false
}
