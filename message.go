package main

import "fmt"

type Message struct {
	SenderID    int
	MessageType int
	Content     interface{}
}

func (self *Message) toString() string {
	msgType := ""
	switch self.MessageType {
	case 0:
		msgType = "Creation"
	case 1:
		msgType = "Acknowledgement"
	case 2:
		msgType = "Renaming"
	default:
		msgType = "Other"
	}
	return fmt.Sprintf("\nSender: %d,\nMessage Type: %d (%s),\nContent: %v\n",
		self.SenderID, self.MessageType, msgType, self.Content)
}
