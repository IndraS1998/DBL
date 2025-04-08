package state

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "raft/raft"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// beggin election triggers new election
func BegginElection(n *Node) {
	mu := sync.Mutex{}
	fmt.Printf("%v Starting election...", n.Address)

	mu.Lock()
	// increment term
	n.CurrentTerm++
	// change state to candidate
	n.Status = "candidate"
	// vote for self
	n.VotedFor = n.Address
	receivedVotes := 1
	mu.Unlock()
	// send request vote to all peers
	c := make(chan bool)
	for _, peer := range n.Peers {
		go performRPC(n, peer, c)
	}
	for i := 0; i < len(n.Peers); i++ {
		tempRes := <-c
		if tempRes {
			receivedVotes++
		}
	}
	if receivedVotes > len(n.Peers)/2 {
		fmt.Printf("received votes %v , %v is now the leader \n", receivedVotes, n.Address)
		mu.Lock()
		n.Status = "leader"
	} else {
		fmt.Printf("received votes %v , %v will revert to follower \n", receivedVotes, n.Address)
	}

	// if vote timer elapse : restart election

	// TODO if receive rpc from a leader : transition back to follower
}

func performRPC(n *Node, peerAddress string, resultChan chan bool) {
	fmt.Printf("sending request vote to %v \n", peerAddress)
	con, err := grpc.NewClient(fmt.Sprintf("localhost:%v", peerAddress), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to server %v:", err)
	}
	defer con.Close()
	c := pb.NewRaftClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	vr, err := c.RequestVote(ctx, &pb.RequestVoteRequest{Term: n.CurrentTerm,
		CandidateId: n.Address, LastLogIndex: n.CommitIndex, LastLogTerm: n.LastApplied})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	// return vr.VoteGranted as a channel
	resultChan <- vr.VoteGranted
}
