package main

import (
	"fmt"
	"strconv"
	"time"
)

var list_processors []*Processor

func StartElection(proc *Processor) {
	proc.annexing_token = Token{
		phase:        0,
		processor_id: proc.ID,
	}
	temp_attachement := Attachment{
		token:        proc.annexing_token,
		hop_counter:  0,
		last_max_hop: -1,
		stack:        make([]chan Attachment, 0),
		visited:      make([]chan Attachment, 0),
	}
	temp_attachement.visited = append(temp_attachement.visited, proc.channel)
	for _, edge := range proc.connected_edges {
		temp_attachement.stack = append(temp_attachement.stack, edge)
	}
	last_index := len(temp_attachement.stack) - 1
	popped_channel := temp_attachement.stack[last_index]
	temp_attachement.stack = temp_attachement.stack[:last_index]
	proc.last_sent_channel = popped_channel
	if leader_elected {
		fmt.Println("Leader already elected, aborting election of proc" + proc.ID)
		return
	}
	fmt.Println("Processor " + proc.ID + " started election")
	popped_channel <- temp_attachement
}

func main() {
	list_processors = make([]*Processor, 0)

	// Create processors
	for i := 0; i < 10; i++ {
		list_processors = append(list_processors, CreateProcessor(strconv.Itoa(i)))
	}

	list_processors[0].connected_edges = append(list_processors[0].connected_edges, list_processors[4].channel)
	list_processors[1].connected_edges = append(list_processors[1].connected_edges, list_processors[4].channel)
	list_processors[2].connected_edges = append(list_processors[2].connected_edges, list_processors[4].channel)
	list_processors[3].connected_edges = append(list_processors[3].connected_edges, list_processors[7].channel)
	list_processors[4].connected_edges = append(list_processors[4].connected_edges, list_processors[0].channel, list_processors[1].channel, list_processors[2].channel, list_processors[5].channel)
	list_processors[5].connected_edges = append(list_processors[5].connected_edges, list_processors[4].channel, list_processors[6].channel)
	list_processors[6].connected_edges = append(list_processors[6].connected_edges, list_processors[5].channel, list_processors[7].channel, list_processors[8].channel)
	list_processors[7].connected_edges = append(list_processors[7].connected_edges, list_processors[3].channel, list_processors[6].channel)
	list_processors[8].connected_edges = append(list_processors[8].connected_edges, list_processors[6].channel, list_processors[9].channel)
	list_processors[9].connected_edges = append(list_processors[9].connected_edges, list_processors[8].channel)

	// Start leader election
	go StartElection(list_processors[0])
	go StartElection(list_processors[1])
	go StartElection(list_processors[2])
	go StartElection(list_processors[3])

	time.Sleep(1000 * time.Millisecond)
	for _, proc := range list_processors {
		go proc.Run()
	}

	// Wait for leader election to finish
	for !leader_elected {
	}

	// Print leader
	println("Leader: " + GetLeader().ID + " with phase " + strconv.Itoa(GetLeader().annexing_token.phase))
}
