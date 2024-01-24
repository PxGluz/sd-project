package main

import "fmt"

var leader_elected bool = false

type Processor struct {
	ID                string
	channel           chan Attachment
	connected_edges   []chan Attachment // names of connected edges
	last_sent_channel chan Attachment
	annexing_token    Token
	candidate         Token
	max_hop           int
	chased            int
}

func CheckFinished() bool {
	// TODO: Replace with actual check
	return false
}

func (self *Processor) Run() {
	for {
		select {
		case msg := <-self.channel:
			if !leader_elected {
				if msg.hop_counter != -1 {
					// annexing mode
					if self.annexing_token == msg.token && CheckFinished() {
						// c2
						leader_elected = true
						fmt.Println("Leader elected: ", self.ID)
						break
					} else if msg.token.phase > self.annexing_token.phase ||
						(msg.token == self.annexing_token && self.chased == -1 && self.candidate.processor_id == "") {
						// c1
						self.annexing_token = msg.token
						msg.hop_counter++
						self.max_hop = msg.hop_counter
						self.candidate = Token{
							phase:        -1,
							processor_id: "",
						}
						for _, edge := range self.connected_edges {
							edge <- msg
							self.last_sent_channel = edge
						}
					} else if msg.token.phase < self.annexing_token.phase {
						// c3
						// do nothing
					} else if msg.token.phase == self.candidate.phase {
						// c4
						self.annexing_token = Token{
							phase:        msg.token.phase + 1,
							processor_id: self.ID,
						}
						msg.hop_counter = 0
						self.max_hop = msg.hop_counter
						self.candidate = Token{
							phase:        -1,
							processor_id: "",
						}
						for _, edge := range self.connected_edges {
							edge <- Attachment{
								token:        self.annexing_token,
								hop_counter:  self.max_hop,
								last_max_hop: -1,
							}
							self.last_sent_channel = edge
						}
					} else if msg.token.phase == self.annexing_token.phase && (self.chased == msg.token.phase || self.annexing_token.processor_id > msg.token.processor_id) {
						// c5
						self.candidate = msg.token
					} else if msg.token.phase == self.annexing_token.phase && self.annexing_token.processor_id < msg.token.processor_id && self.chased == -1 {
						// c6
						self.chased = msg.token.phase
						self.last_sent_channel <- Attachment{
							token:        msg.token,
							hop_counter:  -1,
							last_max_hop: self.max_hop,
						}
					}
				} else if msg.last_max_hop != -1 {
					// chasing mode

				}
			}
		}

	}
}

func CreateProcessor(id string) *Processor {
	temp_processor := Processor{
		ID:      id,
		channel: make(chan Attachment),
		annexing_token: Token{
			phase:        -1,
			processor_id: "",
		},
		candidate: Token{
			phase:        -1,
			processor_id: "",
		},
		max_hop: 0,
		chased:  -1,
	}
	go temp_processor.Run()
	return &temp_processor
}
