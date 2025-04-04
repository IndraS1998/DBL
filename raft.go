package main

import (
	"fmt"
	"time"
)

type Node struct {
	currentTerm, votedFor, commitIndex, lastApplied int32
	timer                                           time.Duration
	leaderAddress                                   int32
	status                                          string
	peers                                           []int32
	address                                         int32
}

func NewNode() *Node {
	return &Node{
		currentTerm:   0,
		votedFor:      -1,
		commitIndex:   0,
		lastApplied:   0,
		timer:         randomTimer(),
		leaderAddress: -1,
		status:        "follower",
		peers:         []int32{30, 40},
		address:       10,
	}
}

func (n *Node) printDetails() {
	fmt.Println("Node Details:")
	fmt.Println("Address:", n.address)
	fmt.Println("Current Term:", n.currentTerm)
	fmt.Println("Voted For:", n.votedFor)
	fmt.Println("Commit Index:", n.commitIndex)
	fmt.Println("Last Applied:", n.lastApplied)
	fmt.Println("Timer:", n.timer)
	fmt.Println("Leader Address:", n.leaderAddress)
	fmt.Println("Status:", n.status)
	fmt.Println("Peers:", n.peers)
	fmt.Println("Node Details End")
}
