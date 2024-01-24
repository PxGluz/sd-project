package main

type Message struct {
	token       Token
	hop_counter int
}

type Token struct {
	phase        int
	processor_id string
}
