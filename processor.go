package main

var processors []Processor

type Processor struct {
	ID              string
	connected_edges []string // names of connected edges
	annexed_token   Token
	max_hop         int
	chased          string
}
