package vote

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "raft/raft"
	"raft/state"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedRaftServer
	node *state.Node
}

func NewServer(node *state.Node) *server {
	return &server{node: node}
}

func (s *server) RequestVote(_ context.Context, vr *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	log.Printf("candidate term: %v candidateId: %v \n", vr.Term, vr.CandidateId)
	if vr.GetTerm() > s.node.CurrentTerm && len(s.node.VotedFor) == 0 {
		s.node.Mu.Lock()
		s.node.CurrentTerm = vr.GetTerm()
		s.node.VotedFor = vr.GetCandidateId()
		s.node.Mu.Unlock()
		s.node.PrintDetails()
		return &pb.RequestVoteResponse{Term: s.node.CurrentTerm, VoteGranted: true}, nil
	} else {
		return &pb.RequestVoteResponse{Term: s.node.CurrentTerm, VoteGranted: false}, nil

	}
}

func StartVoteListenServer(node *state.Node, wg *sync.WaitGroup) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", node.Address))
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterRaftServer(grpcServer, NewServer(node))
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
