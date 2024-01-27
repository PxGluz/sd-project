package main

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
				return &proc
			}
		}
	}

	return nil
}

func (self *Processor) Run() {
	for {
		select {
		case msg := <-self.channel:
			if !leader_elected {
				if msg.hop_counter != -1 {
					// annexing mode
					if self.annexing_token == msg.token && CheckFinished(msg.token.processor_id) {
						// a2
						leader_elected = true
						leader_id = msg.token.processor_id
						break
					} else if msg.token.phase > self.annexing_token.phase ||
						(msg.token == self.annexing_token && self.chased == -1 && self.candidate.processor_id == "") {
						// a1
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
						// a3
						// do nothing
					} else if msg.token.phase == self.candidate.phase {
						// a4
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
						// a5
						self.candidate = msg.token
					} else if msg.token.phase == self.annexing_token.phase && self.annexing_token.processor_id < msg.token.processor_id && self.chased == -1 {
						// a6
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
						// do nothing
					} else if self.candidate.phase == msg.token.phase {
						// c3
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
					} else if (msg.token.phase == self.annexing_token.phase && msg.token.processor_id == self.annexing_token.processor_id &&
						(self.chased == msg.token.phase || self.max_hop < msg.last_max_hop)) ||
						((msg.token.phase != self.annexing_token.phase || msg.token.processor_id != self.annexing_token.processor_id) &&
							msg.token.phase > self.annexing_token.phase) {
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
