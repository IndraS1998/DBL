// representing the raft state
package state

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type LogEntry struct {
	Term int32
	Cmd  string
}
type Node struct {
	CurrentTerm, CommitIndex, LastApplied                              int32
	LeaderAddress, Status, Address, VotedFor                           string
	Peers                                                              []string
	Mu                                                                 sync.RWMutex
	ResetTimerChan, StopTimerChan, StartElectionChan, BecomeLeaderChan chan bool
	NextIndex, MatchIndex                                              map[string]int32
	LOG                                                                []LogEntry
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
		BecomeLeaderChan:  make(chan bool),
		NextIndex:         make(map[string]int32),
		MatchIndex:        make(map[string]int32),
		LOG:               make([]LogEntry, 0),
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
				n.BecomeLeaderChan <- true
				timer.Stop()
				return
			}

		}
	}()
}

/*
	func SendHeartbeat(n *Node, peer string) error {
	    conn, err := grpc.Dial(fmt.Sprintf(":%v", peer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	    if err != nil {
	        return fmt.Errorf("failed to connect to peer %v: %v", peer, err)
	    }
	    defer conn.Close()

	    client := pb.NewRaftClient(conn)

	    // Create an empty AppendEntriesRequest as a heartbeat
	    req := &pb.AppendEntriesRequest{
	        Term:         n.CurrentTerm,
	        LeaderId:     n.Address,
	        PrevLogIndex: int32(len(n.LOG) - 1),
	        PrevLogTerm:  n.LOG[len(n.LOG)-1].Term,
	        Entries:      nil, // No entries, just a heartbeat
	        LeaderCommit: n.CommitIndex,
	    }

	    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	    defer cancel()

	    _, err = client.AppendEntries(ctx, req)
	    if err != nil {
	        return fmt.Errorf("failed to send heartbeat to peer %v: %v", peer, err)
	    }

	    return nil
	}
*/
func (n *Node) PrintDetails() {
	fmt.Println("======================================")
	fmt.Printf("Address: %v, currentTerm : %v, Voted For: %v, Commit Index: %v, Last Applied: %v \n",
		n.Address, n.CurrentTerm, n.VotedFor, n.CommitIndex, n.LastApplied)
	fmt.Printf(" Leader Adress : %v, Status: %v, Peers: %v \n",
		n.LeaderAddress, n.Status, n.Peers)
	fmt.Println("=======================================")
}

/*
Example Scenario

    Leader's log: [1,2,3,4,5] (last log index = 5)

    Follower A's log: [1,2,3] (behind)

    Follower B's log: [1,2,4] (diverged at index 3)

Leader s tracking:

    nextIndex = [4, 3] (for Follower A and B, respectively).

    commitIndex = 2 (if entry 2 is the latest committed one).

The leader will:

    Send entries from nextIndex[A] = 4 to Follower A (entries [4,5]).

    Send entries from nextIndex[B] = 3 to Follower B (entry [3], but it will be rejected, so nextIndex[B] is decremented to 2 and retried).

Once a majority (including the leader and at least one follower) has entry 5, the leader advances commitIndex to 5.

*/
