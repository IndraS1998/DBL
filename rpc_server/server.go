package rpc_server

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
	log.Printf("received vote request from candidateId: %v with term term: %v  \n", vr.CandidateId, vr.Term)
	ct, e := s.node.GetCurrentTerm()
	if e != nil {
		log.Printf("could not get current term: %v", e)
		return nil, e
	}
	if vr.GetTerm() > ct {
		er := s.node.SetCurrentTerm(vr.GetTerm())
		if er != nil {
			log.Printf("could not set current term: %v", er)
			return nil, er
		}
		err := s.node.SetVotedFor(vr.GetCandidateId())
		if err != nil {
			log.Printf("could not set voted for: %v", err)
			return nil, err
		}
		s.node.ResetTimerChan <- true
		s.node.PrintDetails()
		return &pb.RequestVoteResponse{Term: ct, VoteGranted: true}, nil
	} else {
		return &pb.RequestVoteResponse{Term: ct, VoteGranted: false}, nil
	}
}

func (s *server) AppendEntries(_ context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	log.Printf("%v received AppendEntries from %v with term %v\n", s.node.Address, req.LeaderId, req.Term)
	ct, e := s.node.GetCurrentTerm()
	if e != nil {
		log.Printf("could not get current term: %v", e)
		return nil, e
	}

	if req.Term < ct {
		return &pb.AppendEntriesResponse{Term: ct, Success: false}, nil
	}

	// Update term and become follower if necessary
	er := s.node.SetCurrentTerm(req.Term)
	if er != nil {
		log.Printf("could not set current term: %v", er)
		return nil, er
	}
	s.node.Mu.Lock()
	s.node.LeaderAddress = req.LeaderId
	s.node.Mu.Unlock()

	// Reset timer because we received a heartbeat
	s.node.ResetTimerChan <- true
	s.node.PrintDetails()
	// You should add log consistency checks here (prevLogIndex, prevLogTerm, etc.)
	// For now, we just accept the entries and return success

	// Append new entries to the log

	uct, e1 := s.node.GetCurrentTerm()
	if e1 != nil {
		log.Printf("could not get current term: %v", e1)
		return nil, e1
	}
	return &pb.AppendEntriesResponse{Term: uct, Success: true}, nil
}

func StartRPCServerListener(node *state.Node, wg *sync.WaitGroup) {
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
