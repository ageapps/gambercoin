package data

import (
	"reflect"

	"github.com/ageapps/gambercoin/pkg/blockchain"
	"github.com/ageapps/gambercoin/pkg/monguer"
)

const (
	// PACKET_SIMPLE type
	PACKET_SIMPLE = "SIMPLE"
	// PACKET_RUMOR type
	PACKET_RUMOR = "RUMOR"
	// PACKET_STATUS type
	PACKET_STATUS = "STATUS"
	// PACKET_PRIVATE type
	PACKET_PRIVATE = "PRIVATE"
	// PACKET_TX type
	PACKET_TX = "TX_PUBLISH"
	// PACKET_BLOCK type
	PACKET_BLOCK = "BLOCK_PUBLISH"
)

// UDPMessage struct
type UDPMessage struct {
	Address string
	Packet  GossipPacket
	Message Message
}

// GossipPacket struct
type GossipPacket struct {
	Simple       *SimpleMessage
	Rumor        *monguer.RumorMessage
	Status       *monguer.StatusPacket
	Private      *PrivateMessage
	TxMessage    *blockchain.TxMessage
	BlockMessage *blockchain.BlockMessage
}

// GetPacketType function
func (packet *GossipPacket) GetPacketType() string {
	types := []string{
		PACKET_SIMPLE,
		PACKET_RUMOR,
		PACKET_STATUS,
		PACKET_PRIVATE,
		PACKET_TX,
		PACKET_BLOCK,
	}
	var values []interface{}
	values = append(values, packet.Simple)
	values = append(values, packet.Rumor)
	values = append(values, packet.Status)
	values = append(values, packet.Private)
	values = append(values, packet.TxMessage)
	values = append(values, packet.BlockMessage)

	notNull := -1

	for index := 0; index < len(values); index++ {
		if !reflect.ValueOf(values[index]).IsNil() {
			//fmt.Printf("YYYYYY %v - %v - %v\n", types[index], notNull, values[index])
			// 2 or more properties where != null
			if notNull >= 0 {
				return ""
			}
			notNull = index
		}
	}
	if notNull >= 0 {
		return types[notNull]
	}
	return ""
	// switch {
	// case packet.Rumor != nil && packet.Status == nil && packet.Simple == nil && packet.Private == nil && packet.DataReply == nil && packet.DataRequest == nil:
	// 	return PACKET_RUMOR
	// case packet.Rumor == nil && packet.Status != nil && packet.Simple == nil && packet.Private == nil && packet.DataReply == nil && packet.DataRequest == nil:
	// 	return PACKET_STATUS
	// case packet.Rumor == nil && packet.Status == nil && packet.Simple != nil && packet.Private == nil && packet.DataReply == nil && packet.DataRequest == nil:
	// 	return PACKET_SIMPLE
	// case packet.Rumor == nil && packet.Status == nil && packet.Simple == nil && packet.Private != nil && packet.DataReply == nil && packet.DataRequest == nil:
	// 	return PACKET_PRIVATE
	// case packet.Rumor == nil && packet.Status == nil && packet.Simple == nil && packet.Private == nil && packet.DataReply != nil && packet.DataRequest == nil:
	// 	return PACKET_DATA_REPLY
	// case packet.Rumor == nil && packet.Status == nil && packet.Simple == nil && packet.Private == nil && packet.DataReply == nil && packet.DataRequest != nil:
	// 	return PACKET_DATA_REQUEST
	// default:
	// 	return ""
	// }
}
