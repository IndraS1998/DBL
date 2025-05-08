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

// requestVoteRPCStub sends a request vote RPC to the given peer address and returns true if the vote is granted
func requestVoteRPCStub(n *Node, peerAddress string, abort context.CancelFunc) bool {
	fmt.Printf("sending request vote to %v \n", peerAddress)
	con, err := grpc.NewClient(fmt.Sprintf("localhost:%s", peerAddress), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to server %v:", err)
	}
	defer con.Close()
	c := pb.NewRaftClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ct, e := n.Log.GetCurrentTerm()
	if e != nil {
		log.Printf("could not get current term: %v", err)
		return false
	}
	vr, err := c.RequestVote(ctx, &pb.RequestVoteRequest{Term: ct,
		CandidateId: n.Address, LastLogIndex: n.CommitIndex, LastLogTerm: n.LastApplied})
	if err != nil {
		log.Printf("could not greet: %v", err)
		return false
	}
	if vr.Term >= ct {
		log.Printf("node %v has a higher term %v than %v", peerAddress, vr.Term, ct)
		abort()
	}
	return vr.VoteGranted
}

// SendHeartbeat sends a heartbeat to a peer and returns true based on the response of the peer
func appendEntryRPCStub(node *Node, peer string, entrySlice []LogEntry, ct, prevLogIndex, prevLogTerm int32) (*pb.AppendEntriesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	con, err := grpc.NewClient(fmt.Sprintf("localhost:%s", peer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	client := pb.NewRaftClient(con)

	// for each entry received, extract the term, refTable and payload
	protoEntries := make([]*pb.LogEntry, len(entrySlice))
	for i, entry := range entrySlice {
		if protoEntry, err := ToProtoLogEntry(entry, node.Log.DB); err == nil {
			protoEntries[i] = protoEntry
		}
	}
	req := &pb.AppendEntriesRequest{
		Term:         ct,
		LeaderId:     node.Address,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      protoEntries, // empty for heartbeat
		LeaderCommit: int32(node.CommitIndex),
	}

	resp, err := client.AppendEntries(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
