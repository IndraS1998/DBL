// representing the raft state
package state

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"raft/custom_test"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Node struct {
	CommitIndex, LastApplied                                                                 int32
	LeaderAddress, Status, Address                                                           string
	Peers                                                                                    []string
	Mu                                                                                       sync.RWMutex
	ResetTimerChan, StopTimerChan, StartElectionChan, BecomeLeaderChan, RevertToFollowerChan chan bool
	MatchIndex                                                                               map[string]int32
	NextIndex                                                                                map[string]int64
	*PersistentState
}

// creates a new computational node
func NewNode(address string, allPeers []string) *Node {
	peers := make([]string, 0)
	for _, val := range allPeers {
		if val != address {
			peers = append(peers, val)
		}
	}
	ps, err := InitPersistentState(fmt.Sprintf("%s.db", address))
	if err != nil {
		fmt.Println("Error initializing persistent state:", err)
		return nil
	}
	return &Node{
		CommitIndex:          0,
		LastApplied:          0,
		LeaderAddress:        "",
		Status:               "follower",
		Peers:                peers,
		Address:              address,
		ResetTimerChan:       make(chan bool, 1),
		StopTimerChan:        make(chan bool, 1),
		StartElectionChan:    make(chan bool, 1),
		BecomeLeaderChan:     make(chan bool, 1),
		RevertToFollowerChan: make(chan bool, 1),
		NextIndex:            make(map[string]int64),
		MatchIndex:           make(map[string]int32),
		PersistentState:      ps,
	}
}

// StartTimer starts a timer for the node and waits for a timeout to start election or an RPC to reset the timer
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

// BegginElection is called when a node times out and starts an election
func (n *Node) BegginElection() {

	t, err := n.GetCurrentTerm()
	if err != nil {
		fmt.Println("Error getting current term:", err)
		return
	}
	err = n.SetCurrentTerm(t + 1)
	if err != nil {
		fmt.Println("Error setting current term:", err)
		return
	}
	err = n.SetVotedFor(n.Address)
	if err != nil {
		fmt.Println("Error setting vote:", err)
		return
	}
	n.Mu.Lock()
	receivedVotes := 1
	n.Mu.Unlock()
	// send request vote to all peers

	c := make(chan bool, len(n.Peers))
	/*
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
	*/
	for _, peer := range n.Peers {
		go func(p string) {
			VoteGranted := requestVoteRPCStub(n, p)
			c <- VoteGranted
		}(peer)
	}

	for i := 0; i < len(n.Peers); i++ {
		granted := <-c
		if granted {
			receivedVotes++
		}
	}
	if receivedVotes > len(n.Peers)/2 {
		fmt.Printf("received %v votes , %v is now the leader \n", receivedVotes, n.Address)
		logCount, e := n.GetLogLengh()
		if e != nil {
			fmt.Println("Error getting log length:", e.Error())
			return
		}
		n.Mu.Lock()
		n.Status = "leader"
		n.LeaderAddress = n.Address
		for _, peer := range n.Peers {
			n.NextIndex[peer] = logCount
			n.MatchIndex[peer] = 0
		}
		n.Mu.Unlock()
		n.StopTimerChan <- true
	} else {
		n.Mu.Lock()
		n.Status = "follower"
		n.Mu.Unlock()
		n.ResetTimerChan <- true
		fmt.Printf("received votes %v , %v will revert to follower \n", receivedVotes, n.Address)
	}
}

// AppendEntry sends a heartbeat/appendEntry RPCs to all peers and waits for a majority to respond
func (node *Node) AppendEntry() (bool, error) {
	ch := make(chan bool, len(node.Peers))
	responses := int32(1)

	// generate random commands
	cmd := make([]string, 0)
	for i := 0; i < 2; i++ {
		cmd = append(cmd, custom_test.GenerateMessage())
	}

	ct, e := node.GetCurrentTerm()
	if e != nil {
		log.Printf("could not get current term: %v", e)
		return false, e
	}

	// get last log entry
	lastLog, e1 := node.GetLastLogEntry()
	emptyLog := false
	if e1 != nil {
		if errors.Is(e1, gorm.ErrRecordNotFound) {
			emptyLog = true
		} else {
			log.Printf("could not get last log entry: %v", e1)
			return false, e1
		}
	}
	prevLogIndex := 0
	prevLogTerm := int32(0)

	if !emptyLog {
		prevLogIndex = lastLog.Index
		prevLogTerm = lastLog.Term
	}

	// append to personal log
	go func(cmd []string) {
		for _, c := range cmd {
			err := node.AppendLogEntry(ct, c)
			if err != nil {
				log.Printf("could not append log entry: %v", err)
				return
			}
		}
	}(cmd)

	// send append entries to other peers
	for _, peer := range node.Peers {
		go func(p string) {
			res, _ := appendEntryRPCStub(node, p, cmd, ct, int32(prevLogIndex), prevLogTerm)
			ch <- res.Success
		}(peer)
	}

	for i := 0; i < len(node.Peers); i++ {
		granted := <-ch
		if granted {
			responses++
		}
	}
	return responses > int32(len(node.Peers)/2), nil
}

func (n *Node) PrintDetails() {
	ct, err := n.GetCurrentTerm()
	if err != nil {
		fmt.Println("Error getting current term:", err)
		return
	}
	vf, err1 := n.GetVotedFor()
	if err1 != nil {
		fmt.Println("Error getting voted for:", err1)
		return
	}
	fmt.Println("======================================")
	fmt.Printf("Address: %v, currentTerm : %v, Voted For: %v, Commit Index: %v, Last Applied: %v \n",
		n.Address, ct, vf, n.CommitIndex, n.LastApplied)
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
