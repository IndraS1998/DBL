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
func requestVoteRPCStub(n *Node, peerAddress string, close context.CancelFunc) bool {
	fmt.Printf("sending request vote to %v \n", peerAddress)
	con, err := grpc.NewClient(fmt.Sprintf("localhost:%v", peerAddress), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to server %v:", err)
	}
	defer con.Close()
	c := pb.NewRaftClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ct, e := n.GetCurrentTerm()
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
		close()
	}
	return vr.VoteGranted
}

// SendHeartbeat sends a heartbeat to a peer and returns true based on the response of the peer
func appendEntryRPCStub(node *Node, peer string, cmd []string, ct, prevLogIndex, prevLogTerm int32) (*pb.AppendEntriesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.Dial("localhost:"+peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)

	entries := make([]*pb.LogEntry, 0)
	for _, c := range cmd {
		entries = append(entries, &pb.LogEntry{Term: ct, Command: c})
	}

	req := &pb.AppendEntriesRequest{
		Term:         ct,
		LeaderId:     node.Address,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries, // empty for heartbeat
		LeaderCommit: int32(node.CommitIndex),
	}

	resp, err := client.AppendEntries(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
