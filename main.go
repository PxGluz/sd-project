package main

import (
	"fmt"
	"sync"
	"time"
)

type Processor struct {
	ID            int
	ReservedNames map[int]bool
	IncomingMsg   chan Message
	OutgoingMsg   chan Message
	Done          chan bool
}

type Message struct {
	SenderID int
	Name     int
}

var (
	mutex sync.Mutex
	wg    sync.WaitGroup
)

func NewProcessor(id int, reservedNames map[int]bool, incoming chan Message, outgoing chan Message, done chan bool) *Processor {
	return &Processor{
		ID:            id,
		ReservedNames: reservedNames,
		IncomingMsg:   incoming,
		OutgoingMsg:   outgoing,
		Done:          done,
	}
}

func (p *Processor) Run() {
	defer wg.Done()

	// Initializare
	wg.Done()

	for {
		select {
		case msg := <-p.IncomingMsg:
			// Procesare mesaj
			if _, exists := p.ReservedNames[msg.Name]; !exists {
				p.ReservedNames[msg.Name] = true
				fmt.Printf("Processor %d reserved name %d\n", p.ID, msg.Name)
			}
			// Alte cazuri selectate pentru alte evenimente
		}
	}
}

func main() {
	numProcessors := 5
	reservedNames := make(map[int]bool)

	incomingChannels := make([]chan Message, numProcessors)
	outgoingChannels := make([]chan Message, numProcessors)
	doneChannels := make([]chan bool, numProcessors)

	// Creare canale
	for i := 0; i < numProcessors; i++ {
		incomingChannels[i] = make(chan Message)
		outgoingChannels[i] = make(chan Message)
		doneChannels[i] = make(chan bool)
	}

	// Creare procesoare
	processors := make([]*Processor, numProcessors)
	for i := 0; i < numProcessors; i++ {
		processors[i] = NewProcessor(i, reservedNames, incomingChannels[i], outgoingChannels[i], doneChannels[i])
	}

	// Pornirea procesoarelor
	for _, processor := range processors {
		wg.Add(2)
		go processor.Run()
	}

	// Inițierea comunicațiilor între procesori
	for i := 0; i < numProcessors; i++ {
		for j := 0; j < numProcessors; j++ {
			if i != j {
				go func(senderID, receiverID int) {
					for {
						// Simulare generare de mesaje
						time.Sleep(time.Millisecond * 500)
						msg := Message{
							SenderID: senderID,
							Name:     i*10 + senderID, // Simulare alegere de nume unic
						}
						outgoingChannels[receiverID] <- msg
					}
				}(i, j)
			}
		}
	}

	// Așteptarea terminării procesoarelor
	wg.Wait()
}
