package rpc_server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	pb "raft/raft"
	"raft/state"
	"raft/utils"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type server struct {
	pb.UnimplementedRaftServer
	node *state.Node
}

func NewServer(node *state.Node) *server {
	return &server{node: node}
}

func (s *server) RequestVote(_ context.Context, vr *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	ct, e := s.node.Log.GetCurrentTerm()
	if e != nil {
		log.Printf("could not get current term: %v", e)
		return nil, e
	}
	s.node.ResetTimerChan <- true
	if vr.GetTerm() > ct {
		er := s.node.Log.SetCurrentTerm(vr.GetTerm())
		if er != nil {
			log.Printf("could not set current term: %v", er)
			return nil, er
		}
		err := s.node.Log.SetVotedFor(vr.GetCandidateId())
		if err != nil {
			log.Printf("could not set voted for: %v", err)
			return nil, err
		}
		s.node.PrintDetails()
		return &pb.RequestVoteResponse{Term: ct, VoteGranted: true}, nil
	} else {
		return &pb.RequestVoteResponse{Term: ct, VoteGranted: false}, nil
	}
}

func (s *server) AppendEntries(_ context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	s.node.ResetTimerChan <- true

	// Check if the term is less than the current term
	ct, e := s.node.Log.GetCurrentTerm()
	if e != nil {
		log.Printf("could not get current term: %v", e)
		return nil, e
	}
	if req.Term < ct {
		return &pb.AppendEntriesResponse{Term: ct, Success: false}, nil
	}

	// validating prevLogIndex and term

	// if prevLogIndex = 0? i still need to check the log lengh and ensure it is also 0
	if req.PrevLogIndex == 0 {
		logLength, _ := s.node.Log.GetLogLength()
		if logLength != 0 {
			if err := s.node.Log.DeleteLogEntriesFrom(1); err != nil {
				return nil, err
			}
		}
	} else {
		// get log entry at point prevLogIndex
		logEntry, err := s.node.Log.GetLogEntry(int(req.PrevLogIndex))
		// if there is no entry at that point, return false [hint use gorm.RecordNotFound]
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("no such record exists with the index %v\n", req.PrevLogIndex)
			ent, e2 := s.node.Log.GetAllLogEntries()
			if e2 != nil {
				log.Printf("could not get all log entries: %v", e2)
				return nil, e2
			}
			log.Printf("log entries after append for %v: %v", s.node.Address, ent)

			return &pb.AppendEntriesResponse{Term: ct, Success: false}, nil
		}
		// if there is an entry at that index but its term does not match prevLogTerm, return false
		if logEntry.Term != req.PrevLogTerm {
			log.Printf("the entry at index : %v, has term: %v but term : %v was provided", req.PrevLogIndex, logEntry.Term, req.PrevLogTerm)
			return &pb.AppendEntriesResponse{Term: ct, Success: false}, nil
		} else {
			// if previous logs match i need to delete any eventual next log
			if _, err := s.node.Log.GetLogEntry(int(req.PrevLogIndex) + 1); err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					if delEr := s.node.Log.DeleteLogEntriesFrom(int(req.PrevLogIndex) + 1); delEr != nil {
						return nil, delEr
					}
				} else {
					fmt.Println("prev log match and there is no next log. Everything ok")
				}
			}
		}
	}

	// Update term and become follower if necessary
	er := s.node.Log.SetCurrentTerm(req.Term)
	if er != nil {
		log.Printf("could not set current term: %v", er)
		return nil, er
	}
	s.node.Mu.Lock()
	s.node.LeaderAddress = req.LeaderId
	s.node.Mu.Unlock()

	// Append new entries to the log
	payloads := []utils.Payload{}
	for _, entry := range req.Entries {
		payload, err := state.ProtoToLogEntry(entry, entry.ReferenceTable, entry.Term)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}

	err := s.node.Log.AppendLogEntry(payloads)
	if err != nil {
		log.Printf("could not insert log entry: %v", err)
		return nil, err
	}

	// Update commit index
	if req.LeaderCommit > s.node.CommitIndex {
		logLength, _ := s.node.Log.GetLogLength()
		s.node.Mu.Lock()
		s.node.CommitIndex = min(req.LeaderCommit, int32(logLength))
		s.node.Mu.Unlock()
		s.node.Commit() // commit entries here by comparing last applied with actual commit
	}
	s.node.PrintDetails()
	// Print log entries after appending just for testing purposes
	ent, e2 := s.node.Log.GetAllLogEntries()
	if e2 != nil {
		log.Printf("could not get all log entries: %v", e2)
		return nil, e2
	}
	fmt.Printf("log entries after append for %v: %v", s.node.Address, ent)
	uct, e1 := s.node.Log.GetCurrentTerm()
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
