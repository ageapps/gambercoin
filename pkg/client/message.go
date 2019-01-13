package client

// Message to send
type Message struct {
	Text        string
	Destination string
	Broadcast   bool
}

// IsDirectMessage check if is private message
func (msg *Message) IsDirectMessage() bool {
	return msg.Destination != ""
}
