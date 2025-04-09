// representing the raft state
package state

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Node struct {
	CurrentTerm, CommitIndex, LastApplied            int32
	LeaderAddress, Status, Address, VotedFor         string
	Peers                                            []string
	Mu                                               sync.RWMutex
	ResetTimerChan, StopTimerChan, StartElectionChan chan bool
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
		CurrentTerm:       0,
		VotedFor:          "",
		CommitIndex:       0,
		LastApplied:       0,
		LeaderAddress:     "",
		Status:            "follower",
		Peers:             peers,
		Address:           address,
		ResetTimerChan:    make(chan bool),
		StopTimerChan:     make(chan bool),
		StartElectionChan: make(chan bool),
	}
}

// election timer thread #1, starts election when timer elapses
func (n *Node) StartTimer(wg *sync.WaitGroup) {
	go func() {
		for {
			timeout := time.Duration(rand.Int31n(15)+15) * time.Second
			timer := time.NewTimer(timeout)
			fmt.Printf("%v has set a timer for %v seconds \n", n.Address, timeout.Seconds())
			select {
			case <-timer.C:
				fmt.Printf("%v has timed out, starting election \n", n.Address)
				n.StartElectionChan <- true
			case <-n.ResetTimerChan:
				fmt.Printf("%v has received an RPC, restarting timer \n", n.Address)
				if !timer.Stop() {
					<-timer.C
				}
				continue
			case <-n.StopTimerChan:
				fmt.Printf("%v has beacome a leader, stoping global timer \n", n.Address)
				timer.Stop()
				return
			}

		}
	}()
	//BegginElection(n)
}

func (n *Node) PrintDetails() {
	fmt.Println("======================================")
	fmt.Printf("Address: %v, currentTerm : %v, Voted For: %v, Commit Index: %v, Last Applied: %v \n",
		n.Address, n.CurrentTerm, n.VotedFor, n.CommitIndex, n.LastApplied)
	fmt.Printf(" Leader Adress : %v, Status: %v, Peers: %v \n",
		n.LeaderAddress, n.Status, n.Peers)
	fmt.Println("=======================================")
}
