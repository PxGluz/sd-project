package main

type Attachment struct {
	token        Token
	hop_counter  int
	last_max_hop int
	stack        []chan Attachment
	visited      []chan Attachment
}

type Token struct {
	phase        int
	processor_id string
}
