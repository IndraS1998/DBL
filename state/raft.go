// representing the raft state
package state

import (
	"fmt"
	"sync"
	"time"
)

type Node struct {
	CurrentTerm, CommitIndex, LastApplied    int32
	Timer                                    time.Duration
	LeaderAddress, Status, Address, VotedFor string
	Peers                                    []string
}

// creates a new computational node
func NewNode(address string, allPeers []string) *Node {
	peers := make([]string, 0)
	for _, val := range allPeers {
		if val != address {
			peers = append(peers, val)
		}
	}
	return &Node{
		CurrentTerm:   0,
		VotedFor:      "",
		CommitIndex:   0,
		LastApplied:   0,
		Timer:         RandomTimer(),
		LeaderAddress: "",
		Status:        "follower",
		Peers:         peers,
		Address:       address,
	}
}

// election timer thread #1, starts election when timer elapses
func (n *Node) StartTimer(wg *sync.WaitGroup) {
	Countdown(n)
	// beggin vote timer
	BegginElection(n)
}

func (n *Node) PrintDetails() {
	fmt.Println("======================================")
	fmt.Printf("Address: %v, currentTerm : %v, Voted For: %v, Commit Index: %v, Last Applied: %v \n",
		n.Address, n.CurrentTerm, n.VotedFor, n.CommitIndex, n.LastApplied)
	fmt.Printf("Timer: %v, Leader Adress : %v, Status: %v, Peers: %v \n",
		n.Timer, n.LeaderAddress, n.Status, n.Peers)
	fmt.Println("=======================================")
}
