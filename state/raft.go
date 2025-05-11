// representing the raft state
package state

import (
	//"errors"
	"context"
	"fmt"
	"log"
	"math/rand"
	"raft/state/stateMachine"
	"raft/utils"
	"sync"
	"time"
)

type Node struct {
	CommitIndex, LastApplied                                                                 int32
	LeaderAddress, Status, Address                                                           string
	Peers                                                                                    []string
	Mu                                                                                       sync.RWMutex
	ResetTimerChan, StopTimerChan, StartElectionChan, BecomeLeaderChan, RevertToFollowerChan chan bool
	MatchIndex                                                                               map[string]int32
	NextIndex                                                                                map[string]int64
	Log                                                                                      *PersistentState
	StateMachine                                                                             *stateMachine.StateMachine
}

// creates a new computational node
func NewNode(address string, allPeers []string) (*Node, error) {
	peers := make([]string, 0)
	for _, val := range allPeers {
		if val != address {
			peers = append(peers, val)
		}
	}
	ps, err := InitPersistentState(fmt.Sprintf("%s.db", address))
	if err != nil {
		fmt.Println("Error initializing persistent state:", err)
		return nil, fmt.Errorf("could not initialize persistent state for %s, error: %w", address, err)
	}
	sm, sm_init_err := stateMachine.InitStateMachine(fmt.Sprintf("%s.sm.db", address))
	if sm_init_err != nil {
		fmt.Println("Error initializing state machine:", sm_init_err)
		return nil, fmt.Errorf("could not initialize state machine %s, error: %w", address, sm_init_err)
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
		Log:                  ps,
		StateMachine:         sm,
	}, nil
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
				timer.Stop()
				n.StartElectionChan <- true
			case <-n.ResetTimerChan:
				fmt.Printf("%v has received an RPC, restarting timer \n", n.Address)
				if !timer.Stop() {
					<-timer.C
				}
				continue
			case <-n.StopTimerChan:
				fmt.Printf("%v has become a leader, stoping global timer \n", n.Address)
				n.BecomeLeaderChan <- true
				timer.Stop()
				return
			}

		}
	}()
}

// BegginElection is called when a node times out and starts an election
func (n *Node) BeginElection() {

	t, err := n.Log.GetCurrentTerm()
	if err != nil {
		fmt.Println("Error getting current term:", err)
		return
	}
	err = n.Log.SetCurrentTerm(t + 1)
	if err != nil {
		fmt.Println("Error setting current term:", err)
		return
	}
	err = n.Log.SetVotedFor(n.Address)
	if err != nil {
		fmt.Println("Error setting vote:", err)
		return
	}
	mu := sync.Mutex{}
	receivedVotes := 1

	c := make(chan bool, len(n.Peers))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for _, peer := range n.Peers {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			VoteGranted := requestVoteRPCStub(n, p, cancel)
			c <- VoteGranted
		}(peer)
	}

	// launch another go routine to wait for when all the wg have called Done to close the chan
	go func() {
		wg.Wait()
		close(c)
	}()

	for {
		select {
		case granted, open := <-c:
			if !open {
				if receivedVotes > len(n.Peers)/2 {
					fmt.Printf("received %v votes , %v is now the leader \n", receivedVotes, n.Address)
					logCount, e := n.Log.GetLogLength()
					if e != nil {
						fmt.Println("Error getting log length:", e.Error())
						return
					}
					n.Mu.Lock()
					n.Status = "leader"
					n.LeaderAddress = n.Address
					for _, peer := range n.Peers {
						n.NextIndex[peer] = logCount + 1
						n.MatchIndex[peer] = 0
					}
					n.Mu.Unlock()
					n.StopTimerChan <- true
					return
				} else {
					n.Mu.Lock()
					n.Status = "follower"
					n.Mu.Unlock()
					n.ResetTimerChan <- true
					fmt.Printf("received votes %v , %v will revert to follower \n", receivedVotes, n.Address)
					return
				}
			} else if granted {
				mu.Lock()
				receivedVotes++
				mu.Unlock()
			}
		case <-ctx.Done():
			fmt.Println("some node had a higher term!")
			n.RevertToFollowerChan <- true
			return
		}
	}
}

// AppendEntry sends a heartbeat/appendEntry RPCs to all peers and waits for a majority to respond
func (node *Node) AppendEntry() {
	ch := make(chan bool, len(node.Peers))
	responses := int32(1)

	// Generate and append commands to leader's log
	requests, err := utils.RetreivePayloads()
	if err != nil {
		requests = []utils.Payload{}
		fmt.Println(err)
	}
	// get the term
	ct, err := node.Log.GetCurrentTerm()
	if err != nil {
		log.Printf("could not get current term: %v", err)
		return
	}

	// append operations to log
	for _, payload := range requests {
		if err := node.Log.AppendLogEntry(ct, payload); err != nil {
			log.Printf("could not append log entry: %v", err)
			return
		}
	}
	// just for testing purposes
	ent, e2 := node.Log.GetAllLogEntries()
	if e2 != nil {
		log.Printf("could not get all log entries: %v\n", e2)
		return
	}
	log.Printf("log entries after append for %v: %v \n", node.Address, ent)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	for _, peer := range node.Peers {
		wg.Add(1)
		go func(peer string) {
			defer wg.Done()

			prevIndex := int32(node.NextIndex[peer] - 1)
			prevTerm := int32(0)
			//adjust the prevTerm based on the prevIndex
			if prevIndex > 0 {
				entry, err := node.Log.GetLogEntry(int(prevIndex))
				if err != nil {
					log.Printf("could not get log entry: %v", err)

				}
				prevTerm = entry.Term
			}

			//fetch the actual commands based on the last commited entryso as to make the node be up to date
			entries, err := node.Log.GetCommandsFromIndex(int(node.NextIndex[peer]))
			if err != nil {
				log.Printf("could not get commands from index: %v", err)
			}
			res, _ := appendEntryRPCStub(node, peer, entries, ct, prevIndex, prevTerm)
			if res.Success {
				node.Mu.Lock()
				node.NextIndex[peer] += int64(len(entries))
				node.MatchIndex[peer] += int32(len(entries))
				node.Mu.Unlock()
				ch <- true
			} else if res.Term > ct {
				cancel() // Trigger early exit due to stale term
			} else {
				node.Mu.Lock()
				node.NextIndex[peer]--
				node.Mu.Unlock()
				ch <- false
			}
		}(peer)
	}

	// launch another go routine to wait for when all the wg have called Done to close the chan
	go func() {
		wg.Wait()
		close(ch)
	}()

	// evalution is done here
	for {
		select {
		case granted, open := <-ch:
			if !open {
				// All goroutines are done
				if responses > int32(len(node.Peers)/2) {
					lastEntry, _ := node.Log.GetLastLogEntry()
					node.Mu.Lock()
					node.CommitIndex = int32(lastEntry.Index)
					node.Mu.Unlock()
					node.Commit()
					node.PrintDetails()
					return
				} else {
					return
				}
			}
			if granted {
				responses++
			}
		case <-ctx.Done():
			fmt.Println("some node had a higher term!")
			node.RevertToFollowerChan <- true
		}
	}
}

func (n *Node) Commit() {
	if n.CommitIndex > n.LastApplied {
		// now fetch all entries that fall in the range of last applied but less than commit index
		entries, err := n.Log.GetEntriesForCommit(int(n.LastApplied), int(n.CommitIndex))
		if err != nil {
			fmt.Println(err)
		}
		n.Mu.Lock()
		defer n.Mu.Unlock()
		for _, entry := range entries {
			// cast the payload to the correct type
			if entry.Status == utils.TxSuccess {
				fmt.Println("Already applied:")
				n.LastApplied++
				continue
			} else {
				switch entry.ReferenceTable {
				case utils.RefUser:

					var payload UserPayload
					if err := n.Log.DB.First(&payload, entry.PayloadID).Error; err != nil {
						fmt.Printf("failed to load user payload:")
						n.Log.DB.Model(&entry).Update("status", utils.TxFailed)
						return
					}
					userPayload := utils.UserPayload{
						FirstName:                *payload.FirstName,
						LastName:                 *payload.LastName,
						HashedPassword:           *payload.HashedPassword,
						Email:                    *payload.Email,
						DateOfBirth:              *payload.DateOfBirth,
						IdentificationNumber:     *payload.IdentificationNumber,
						IdentificationImageFront: *payload.IdentificationImageFront,
						IdentificationImageBack:  *payload.IdentificationImageBack,
						PrevPW:                   *payload.PrevPW,
						NewPW:                    *payload.NewPW,
						UserID:                   *payload.UserID,
						Action:                   payload.Action,
					}
					if err2 := n.StateMachine.ApplyUserOperation(userPayload); err2 != nil {
						fmt.Printf("failed to apply user operation: %v", err2)
						n.Log.DB.Model(&entry).Update("status", utils.TxFailed)
						return
					}

				case utils.RefAdmin:

					var payload AdminPayload
					if err := n.Log.DB.First(&payload, entry.PayloadID).Error; err != nil {
						fmt.Printf("failed to load user payload:")
						n.Log.DB.Model(&entry).Update("status", utils.TxFailed)
						return
					}
					adminPayload := utils.AdminPayload{
						FirstName:      *payload.FirstName,
						LastName:       *payload.LastName,
						HashedPassword: *payload.HashedPassword,
						Email:          *payload.Email,
						AdminID:        *payload.AdminID,
						UserId:         *payload.UserId,
						Action:         payload.Action,
					}
					if err2 := n.StateMachine.ApplyAdminOperations(adminPayload); err2 != nil {
						fmt.Printf("failed to apply entry to state machine: %v", err2)
						n.Log.DB.Model(&entry).Update("status", utils.TxFailed)
						return
					}

				case utils.RefWallet:

					var payload WalletOperationPayload
					if err := n.Log.DB.First(&payload, entry.PayloadID).Error; err != nil {
						fmt.Printf("failed to load wallet payload")
						n.Log.DB.Model(&entry).Update("status", utils.TxFailed)
						return
					}
					walletPayload := utils.WalletOperationPayload{
						Wallet1: payload.Wallet1,
						Wallet2: *payload.Wallet2,
						Amount:  payload.Amount,
						Action:  payload.Action,
					}
					if err2 := n.StateMachine.ApplyWalletOperation(walletPayload); err2 != nil {
						fmt.Printf("failed to apply entry to state machine: %v", err2)
						n.Log.DB.Model(&entry).Update("status", utils.TxFailed)
						return
					}

				default:
					fmt.Println("Unknown reference table")
					n.Log.DB.Model(&entry).Update("status", utils.TxFailed)
				}
				//TODO : this should be in some transaction format
				n.LastApplied++
				n.Log.DB.Model(&entry).Update("status", utils.TxSuccess)
			}
		}
	} else {
		fmt.Println("Nothing to commit")
	}
}

func (n *Node) PrintDetails() {
	ct, err := n.Log.GetCurrentTerm()
	if err != nil {
		fmt.Println("Error getting current term:", err)
		return
	}
	vf, err1 := n.Log.GetVotedFor()
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
