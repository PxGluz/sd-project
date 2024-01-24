package main

type Attachment struct {
	token        Token
	hop_counter  int
	last_max_hop int
}

type Token struct {
	phase        int
	processor_id string
}
