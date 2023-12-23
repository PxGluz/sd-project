package main

import (
	"errors"
	"fmt"
)

var processors []Processor

type Processor struct {
	ID             int
	processors_ids []int
	chan_messages  chan Message
}

func NewProcessor(id int) *Processor {
	var temp_processor = Processor{
		ID:             id,
		processors_ids: []int{},
		chan_messages:  make(chan Message),
	}
	for _, processor := range processors {
		processor.chan_messages <- Message{
			SenderID:    temp_processor.ID,
			MessageType: 0,
			Content:     fmt.Sprintf("New Processor: {%d}", temp_processor.ID),
		}
	}
	processors = append(processors, temp_processor)
	return &temp_processor
}

func GetChannelByID(id int) (chan Message, error) {
	for _, processor := range processors {
		if processor.ID == id {
			return processor.chan_messages, nil
		}
	}
	return nil, errors.New("Processor not found!")
}

func (self *Processor) SendMessage(id_receiver int, msg Message) error {
	chan_receiver, err := GetChannelByID(id_receiver)

	if err != nil {
		return err
	}

	chan_receiver <- msg

	if msg.MessageType != 1 {
		self.processors_ids = deleteElement(self.processors_ids, id_receiver)
	}

	return nil
}

func (self *Processor) ReceiveMessageCreation(msg Message) {
	self.processors_ids = append(self.processors_ids, msg.SenderID)

	msg_ack := Message{
		SenderID:    self.ID,
		MessageType: 1,
		Content:     "Creation acknowledged",
	}

	self.SendMessage(msg.SenderID, msg_ack)
}

func (self *Processor) ReceiveMessageAcknowledgement(msg Message) {
	self.processors_ids = append(self.processors_ids, msg.SenderID)
}

func (self *Processor) ReceiveMessageRenaming(msg Message) error {
	self.processors_ids = deleteElement(self.processors_ids, msg.SenderID)
	if new_id, ok := msg.Content.(int); ok {
		self.processors_ids = append(self.processors_ids, new_id)
	} else {
		return errors.New("Invalid name received")
	}
	msg_ack := Message{
		SenderID:    self.ID,
		MessageType: 1,
		Content:     "Renaming acknowledged",
	}

	self.SendMessage(msg.SenderID, msg_ack)
	return nil
}

func (self *Processor) ReceiveMessageDefault(msg Message) {
	fmt.Println(msg.Content)

	msg_ack := Message{
		SenderID:    self.ID,
		MessageType: 1,
		Content:     "Default message acknowledged",
	}

	self.SendMessage(msg.SenderID, msg_ack)
}

func (self *Processor) Run() {
	for {
		select {
		case msg := <-self.chan_messages:
			switch msg.MessageType {
			case 0:
				self.ReceiveMessageCreation(msg)
			case 1:
				self.ReceiveMessageAcknowledgement(msg)
			case 2:
				err := self.ReceiveMessageRenaming(msg)
				if err != nil {
					fmt.Println(err)
				}
			default:
				self.ReceiveMessageDefault(msg)
			}
		}
	}
}

func deleteElement(slice []int, valueToDelete int) []int {
	for i, v := range slice {
		if v == valueToDelete {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
