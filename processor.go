package main

import (
	"fmt"
	"slices"
)

var leader_elected bool = false
var leader_id string = ""

type Processor struct {
	ID                string
	channel           chan Attachment
	connected_edges   []chan Attachment // names of connected edges
	last_sent_channel chan Attachment
	annexing_token    Token
	candidate         Token
	max_hop           int
	chased            int
	leader_id         string
}

func CheckFinished(token_processor_id string) bool {
	for _, proc := range list_processors {
		if proc.annexing_token.processor_id != token_processor_id {
			return false
		}
	}
	return true
}

func GetLeader() *Processor {
	if leader_elected {
		for _, proc := range list_processors {
			if proc.ID == leader_id {
				return proc
			}
		}
	}

	return nil
}

func AnnounceLeader() {
	if leader_elected {
		for _, proc := range list_processors {
			proc.leader_id = leader_id
		}
	}
}

func (self *Processor) Run() {
	for {
		select {
		case msg := <-self.channel:
			fmt.Println("Processor ", self.ID, " received token ", msg.token.processor_id, " with phase ", msg.token.phase)
			if !leader_elected {
				if msg.hop_counter != -1 {
					// annexing mode
					if self.annexing_token == msg.token && CheckFinished(msg.token.processor_id) {
						// a2
						fmt.Println("Leader elected: ", msg.token.processor_id)
						leader_elected = true
						leader_id = msg.token.processor_id
						break
					} else if msg.token.phase > self.annexing_token.phase ||
						(msg.token == self.annexing_token && self.chased == -1 && self.candidate.processor_id == "") {
						// a1
						fmt.Println("Processor ", self.ID, " getting annexed by ", msg.token.processor_id, " with phase ", msg.token.phase)
						self.annexing_token = msg.token
						msg.hop_counter++
						self.max_hop = msg.hop_counter
						self.candidate = Token{
							phase:        -1,
							processor_id: "",
						}
						msg.visited = append(msg.visited, self.channel)
						for _, edge := range self.connected_edges {
							if slices.Contains(msg.visited, edge) {
								continue
							}
							msg.stack = append(msg.stack, edge)
						}
						if len(msg.stack) > 0 {
							last_pos := len(msg.stack) - 1
							popped_channel := msg.stack[last_pos]
							msg.stack = msg.stack[:last_pos]
							self.last_sent_channel = popped_channel
							popped_channel <- msg
						} else if len(msg.visited) == len(list_processors) && CheckFinished(self.annexing_token.processor_id) {
							leader_id = msg.token.processor_id
							AnnounceLeader()
							leader_elected = true
							fmt.Println("Leader elected: ", self.annexing_token.processor_id)
						}

					} else if msg.token.phase < self.annexing_token.phase {
						// a3
						fmt.Println("Processor", self.ID, "doing nothing")
						// do nothing
					} else if msg.token.phase == self.candidate.phase {
						// a4
						fmt.Println("Processor ", self.ID, " starting new annexing phase with phase ", msg.token.phase+1, " and processor id ", self.ID)
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
						temp_attachment := Attachment{
							token:        self.annexing_token,
							hop_counter:  self.max_hop,
							last_max_hop: -1,
							visited:      make([]chan Attachment, 0),
							stack:        make([]chan Attachment, 0),
						}
						temp_attachment.visited = append(temp_attachment.visited, self.channel)
						for _, edge := range self.connected_edges {
							if slices.Contains(temp_attachment.visited, edge) {
								continue
							}
							temp_attachment.stack = append(temp_attachment.stack, edge)
						}
						if len(temp_attachment.stack) > 0 {
							last_pos := len(temp_attachment.stack) - 1
							popped_channel := temp_attachment.stack[last_pos]
							temp_attachment.stack = temp_attachment.stack[:last_pos]
							self.last_sent_channel = popped_channel
							popped_channel <- temp_attachment
						} else if len(temp_attachment.visited) == len(list_processors) && CheckFinished(self.annexing_token.processor_id) {
							leader_id = msg.token.processor_id
							AnnounceLeader()
							leader_elected = true
						}

					} else if msg.token.phase == self.annexing_token.phase && (self.chased == msg.token.phase || self.annexing_token.processor_id > msg.token.processor_id) {
						// a5
						fmt.Println("Processor ", self.ID, " new candidate ", msg.token.processor_id, " with phase ", msg.token.phase)
						self.candidate = msg.token
					} else if msg.token.phase == self.annexing_token.phase && self.annexing_token.processor_id < msg.token.processor_id && self.chased == -1 {
						// a6
						fmt.Println("Token ", msg.token.processor_id, " started chasing ", self.annexing_token.processor_id, " with phase ", msg.token.phase)
						self.chased = msg.token.phase
						self.last_sent_channel <- Attachment{
							token:        msg.token,
							hop_counter:  -1,
							last_max_hop: self.max_hop,
						}
					}
				} else if msg.last_max_hop != -1 {
					// chasing mode
					// c1
					fmt.Println("Processor ", self.ID, " chasing ", msg.token.processor_id, " with phase ", msg.token.phase)
					if msg.token.phase == self.annexing_token.phase && msg.token.processor_id == self.annexing_token.processor_id &&
						msg.last_max_hop < self.max_hop &&
						self.chased == -1 &&
						self.candidate.phase != msg.token.phase {
						self.chased = msg.token.phase
						self.last_sent_channel <- Attachment{
							token:        msg.token,
							hop_counter:  -1,
							last_max_hop: self.max_hop,
						}
					} else if msg.token.phase < self.annexing_token.phase {
						// c2
						fmt.Println("Processor ", self.ID, " chase stopped because ", msg.token.processor_id, " with phase ", msg.token.phase, " is in a lower phase")
						// do nothing
					} else if self.candidate.phase == msg.token.phase {
						// c3
						fmt.Println("Processor ", self.ID, " starting new annexing phase with phase ", msg.token.phase+1, " and processor id ", self.ID)
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
						temp_attachment := Attachment{
							token:        self.annexing_token,
							hop_counter:  self.max_hop,
							last_max_hop: -1,
							visited:      make([]chan Attachment, 0),
							stack:        make([]chan Attachment, 0),
						}
						temp_attachment.visited = append(temp_attachment.visited, self.channel)
						for _, edge := range self.connected_edges {
							if slices.Contains(temp_attachment.visited, edge) {
								continue
							}
							temp_attachment.stack = append(temp_attachment.stack, edge)
						}
						if len(temp_attachment.stack) > 0 {
							last_pos := len(temp_attachment.stack) - 1
							popped_channel := temp_attachment.stack[last_pos]
							temp_attachment.stack = temp_attachment.stack[:last_pos]
							self.last_sent_channel = popped_channel
							popped_channel <- temp_attachment
						} else if len(temp_attachment.visited) == len(list_processors) && CheckFinished(self.annexing_token.processor_id) {
							leader_id = msg.token.processor_id
							AnnounceLeader()
							leader_elected = true
						}
					} else if (msg.token.phase == self.annexing_token.phase && msg.token.processor_id == self.annexing_token.processor_id &&
						(self.chased == msg.token.phase || self.max_hop < msg.last_max_hop)) ||
						((msg.token.phase != self.annexing_token.phase || msg.token.processor_id != self.annexing_token.processor_id) &&
							msg.token.phase > self.annexing_token.phase) {
						fmt.Println("Processor ", self.ID, " new candidate ", msg.token.processor_id, " with phase ", msg.token.phase)
						self.candidate.phase = msg.token.phase
						self.candidate.processor_id = msg.token.processor_id
					}
				}
			}
		}
	}
}

func CreateProcessor(id string) *Processor {
	temp_processor := Processor{
		ID:                id,
		channel:           make(chan Attachment),
		connected_edges:   make([]chan Attachment, 0),
		last_sent_channel: make(chan Attachment),
		annexing_token: Token{
			phase:        -1,
			processor_id: "",
		},
		candidate: Token{
			phase:        -1,
			processor_id: "",
		},
		max_hop:   0,
		chased:    -1,
		leader_id: "",
	}
	return &temp_processor
}
