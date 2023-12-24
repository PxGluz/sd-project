package main

import (
	"errors"
	"fmt"
)

var processors []Processor
var log_count int = 0

type Processor struct {
	ID             int
	processors_ids []int
	chan_messages  chan Message
	chan_failure   chan bool
}

func NewProcessor(id int) *Processor {
	var temp_processor = Processor{
		ID:             id,
		processors_ids: []int{},
		chan_messages:  make(chan Message),
		chan_failure:   make(chan bool),
	}
	msg := Message{
		SenderID:    temp_processor.ID,
		MessageType: 0,
		Content:     fmt.Sprintf("New Processor: {%d}", temp_processor.ID),
	}
	for _, processor := range processors {
		go temp_processor.SendMessage(processor.ID, msg)
	}
	processors = append(processors, temp_processor)
	go temp_processor.Run()
	return &temp_processor
}

func (self *Processor) SendMessage(id_receiver int, msg Message) {
	if id_receiver == self.ID {
		fmt.Println("Can't send message to self")
		return
	}
	chan_receiver, err := getChannelById(id_receiver)

	if msg.MessageType != 1 {
		self.processors_ids = deleteElement(self.processors_ids, id_receiver)
	}

	if err != nil {
		self.LogMessage(err.Error())
		self.processors_ids = deleteElement(self.processors_ids, id_receiver)
	} else {
		chan_receiver <- msg
		self.LogMessage(fmt.Sprintf("Message sent to %d:%s", id_receiver, msg.toString()))
	}
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
			self.LogMessage(fmt.Sprintf("Received message from %d:%s", msg.SenderID, msg.toString()))
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
		case <-self.chan_failure:
			break
		}
	}
}

func (self *Processor) LogMessage(msg string) {
	fmt.Println(fmt.Sprintf("Processor %d: %s", self.ID, msg))
	log_count++
	//fmt.Println(log_count)
}

func deleteElement(slice []int, valueToDelete int) []int {
	for i, v := range slice {
		if v == valueToDelete {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func printProcessorsIds() {
	fmt.Print("Active processors: ")
	for _, processor := range processors {
		fmt.Print(fmt.Sprintf("%d ", processor.ID))
	}
	fmt.Println()
}

func deleteProcessorById(id int) error {
	for i, processor := range processors {
		if processor.ID == id {
			processor.LogMessage("Removed from system. Shutting down...")
			processor.chan_failure <- true
			processors = append(processors[:i], processors[i+1:]...)

			return nil
		}
	}

	return errors.New("Processor not found")
}

func getChannelById(id int) (chan Message, error) {
	for _, processor := range processors {
		if processor.ID == id {
			return processor.chan_messages, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Processor %d not found!", id))
}

func getProcessorById(id int) (*Processor, error) {
	for _, processor := range processors {
		if processor.ID == id {
			return &processor, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Processor %d not found!", id))
}
