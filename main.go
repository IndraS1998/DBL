package main

import (
	"raft/state"
	"raft/vote"
	"sync"
)

func main() {
	peers := []string{"9001", "9002", "9003"}
	wg := sync.WaitGroup{}
	wg.Add(2)
	// create nodes
	n1 := state.NewNode("9001", peers)
	n1.PrintDetails()
	n2 := state.NewNode("9002", peers)
	n2.PrintDetails()
	n3 := state.NewNode("9003", peers)
	n3.PrintDetails()
	/*
		n4 := state.NewNode("9004", peers)
		n4.PrintDetails()
		n5 := state.NewNode("9005", peers)
		n5.PrintDetails()
	*/
	//start election response server
	go vote.StartVoteListenServer(n1, &wg)
	go vote.StartVoteListenServer(n2, &wg)
	go vote.StartVoteListenServer(n3, &wg)
	/*
		go vote.StartVoteListenServer(n4, &wg)
		go vote.StartVoteListenServer(n5, &wg)
	*/
	//start various timers
	n1.StartTimer(&wg)
	n2.StartTimer(&wg)
	n3.StartTimer(&wg)
	/*
		go n4.StartTimer(&wg)
		go n5.StartTimer(&wg)
	*/
	go func() {
		for {
			select {
			case <-n1.StartElectionChan:
				state.BegginElection(n1)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-n1.StartElectionChan:
				state.BegginElection(n2)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-n1.StartElectionChan:
				state.BegginElection(n3)
			}
		}
	}()
	wg.Wait()
}
