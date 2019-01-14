package stack

import (
	"sync"

	"github.com/ageapps/gambercoin/pkg/logger"
	"github.com/ageapps/gambercoin/pkg/monguer"
)

const (
	// OLD_MESSAGE type
	OLD_MESSAGE = "OLD_MESSAGE"
	// NEW_MESSAGE type
	NEW_MESSAGE = "NEW_MESSAGE"
	// IN_SYNC type
	IN_SYNC = "IN_SYNC"
)

// GenericMessage that can be saved in a stack
type GenericMessage interface {
	GetID() uint32
	GetOrigin() string
}

// MessageStack struct
// that contains as keys the origin
// and as value an array of the rumor
// messages received by that origin
type MessageStack struct {
	Messages map[string][]GenericMessage
	sync.Mutex
}

// NewMessageStack func
func NewMessageStack() MessageStack {
	return MessageStack{
		Messages: make(map[string][]GenericMessage),
	}
}

// CompareMessage func
func (stack *MessageStack) CompareMessage(origin string, id uint32) string {
	stack.Lock()
	defer stack.Unlock()
	messages, ok := stack.Messages[origin]
	if !ok || len(messages) <= 0 {
		return NEW_MESSAGE
	}
	lastMessageID := messages[len(messages)-1].GetID()
	logger.Logv("Comparing messages %v/%v", lastMessageID, id)
	switch {
	case id == lastMessageID:
		return IN_SYNC
	case id > lastMessageID:
		return NEW_MESSAGE
	case id < lastMessageID:
		return OLD_MESSAGE
	}
	return ""
}

// GetMessage func
func (stack *MessageStack) GetMessage(origin string, id uint32) *GenericMessage {
	stack.Lock()
	defer stack.Unlock()
	messages, ok := stack.Messages[origin]
	if !ok || len(messages) <= 0 {
		return nil
	}
	for _, msg := range messages {
		if msg.GetID() == id {
			return &msg
		}
	}
	return nil
}

//AddMessage func
func (stack *MessageStack) AddMessage(msg GenericMessage) {
	stack.Lock()
	defer stack.Unlock()
	id := msg.GetID()
	origin := msg.GetOrigin()
	messages, ok := stack.Messages[origin]
	if !ok {
		stack.Messages[origin] = []GenericMessage{msg}
	} else {
		lastMessageID := messages[len(messages)-1].GetID()
		if id == uint32(lastMessageID+1) {
			stack.Messages[origin] = append(messages, msg)
		}
	}
	logger.Logi("Message appended to stack Origin:%v ID:%v", origin, id)
}

// PrintStack func
func (stack *MessageStack) PrintStack() {
	stack.Lock()
	defer stack.Unlock()
	for address := range stack.Messages {
		logger.Logw("Sender <%v>, last message %v", address, stack.Messages[address])
	}
}

// GetStackMap to get a map
// with latest ids saved from each origin
func (stack *MessageStack) GetStackMap() *map[string]uint32 {
	stack.Lock()
	defer stack.Unlock()
	var stackMap = make(map[string]uint32)
	for origin := range stack.Messages {
		stackMap[origin] = stack.getLatestMessageID(origin)
	}
	return &stackMap
}

func (stack *MessageStack) getLatestMessageID(origin string) uint32 {
	messages := stack.Messages[origin]
	lastID := uint32(messages[len(messages)-1].GetID())
	return lastID
}

// GetLatestMessages function
// returns an array with the latest rumor messages
func (stack *MessageStack) GetLatestMessages() *[]GenericMessage {
	stack.Lock()
	defer stack.Unlock()
	var latestMessages = []GenericMessage{}
	for address := range stack.Messages {
		messages := stack.Messages[address]
		latestMessages = append(latestMessages, messages[len(messages)-1])
	}
	return &latestMessages
}

func (stack *MessageStack) GetStatusMessage() *monguer.StatusPacket {
	stack.Lock()
	defer stack.Unlock()
	var vector []monguer.PeerStatus
	for address := range stack.Messages {
		messages := stack.Messages[address]
		peerStatus := monguer.PeerStatus{Identifier: address, NextID: uint32(messages[len(messages)-1].GetID() + 1)}
		vector = append(vector, peerStatus)
	}
	return monguer.NewStatusPacket(&vector, "")
}

// GetStack func
func (stack *MessageStack) GetStack() *map[string][]GenericMessage {
	stack.Lock()
	defer stack.Unlock()
	return &stack.Messages
}

// GetFirstMissingMessage gibben an array of status messages from another peer,
// look for the first message from an origin missimg in the status array
func (stack *MessageStack) GetFirstMissingMessage(comparedMessages *[]monguer.PeerStatus) *GenericMessage {
	for origin, messages := range *stack.GetStack() {
		firstMessage := messages[0]
		found := false
		for _, status := range *comparedMessages {
			if status.Identifier == origin {
				found = true
				break
			}
		}
		if !found {
			logger.Logi("Peer needs to update Origin:%v - ID:%v", origin, firstMessage.GetID())
			return &firstMessage
		}
	}
	return nil
}
