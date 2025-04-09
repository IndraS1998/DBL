package state

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "raft/raft"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// beggin election triggers new election
func BegginElection(n *Node) {

	n.Mu.Lock()
	n.CurrentTerm++
	n.Status = "candidate"
	n.VotedFor = n.Address
	receivedVotes := 1
	n.Mu.Unlock()
	// send request vote to all peers
	c := make(chan bool, len(n.Peers))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for _, peer := range n.Peers {
		go func(p string) {
			VoteGranted := performRPC(n, p)
			c <- VoteGranted
		}(peer)
	}

	for i := 0; i < len(n.Peers); i++ {
		select {
		case granted := <-c:
			if granted {
				receivedVotes++
			}
		case <-ctx.Done():
			fmt.Println("vote collection timedout restarting election")
			n.StartElectionChan <- true
			return
		}
	}
	if receivedVotes > len(n.Peers)/2 {
		fmt.Printf("received %v votes , %v is now the leader \n", receivedVotes, n.Address)
		n.Mu.Lock()
		n.Status = "leader"
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

func performRPC(n *Node, peerAddress string) bool {
	fmt.Printf("sending request vote to %v \n", peerAddress)
	con, err := grpc.NewClient(fmt.Sprintf("localhost:%v", peerAddress), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to server %v:", err)
	}
	defer con.Close()
	c := pb.NewRaftClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	vr, err := c.RequestVote(ctx, &pb.RequestVoteRequest{Term: n.CurrentTerm,
		CandidateId: n.Address, LastLogIndex: n.CommitIndex, LastLogTerm: n.LastApplied})
	if err != nil {
		log.Printf("could not greet: %v", err)
	}
	return vr.VoteGranted
}
