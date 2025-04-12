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
	/*
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
	*/
	for _, peer := range n.Peers {
		go func(p string) {
			VoteGranted := performRPC(n, p)
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
		n.Mu.Lock()
		n.Status = "leader"
		n.LeaderAddress = n.Address
		for _, peer := range n.Peers {
			n.NextIndex[peer] = int32(len(n.LOG))
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

func performRPC(n *Node, peerAddress string) bool {
	fmt.Printf("sending request vote to %v \n", peerAddress)
	con, err := grpc.NewClient(fmt.Sprintf("localhost:%v", peerAddress), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to server %v:", err)
	}
	defer con.Close()
	c := pb.NewRaftClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	vr, err := c.RequestVote(ctx, &pb.RequestVoteRequest{Term: n.CurrentTerm,
		CandidateId: n.Address, LastLogIndex: n.CommitIndex, LastLogTerm: n.LastApplied})
	if err != nil {
		log.Printf("could not greet: %v", err)
	}
	return vr.VoteGranted
}

// SendHeartbeat sends a heartbeat to the specified peer
func SendHeartbeat(node *Node) (bool, error) {
	ch := make(chan bool, len(node.Peers))
	responses := int32(1)
	for _, peer := range node.Peers {
		go func(p string) {
			res, _ := callAppendEntriesRPC(node, p)
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

func callAppendEntriesRPC(node *Node, peer string) (*pb.AppendEntriesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.Dial("localhost:"+peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)

	req := &pb.AppendEntriesRequest{
		Term:         node.CurrentTerm,
		LeaderId:     node.Address,
		PrevLogIndex: 0,                // placeholder for now
		PrevLogTerm:  0,                // placeholder for now
		Entries:      []*pb.LogEntry{}, // empty for heartbeat
		LeaderCommit: int32(node.CommitIndex),
	}

	resp, err := client.AppendEntries(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
