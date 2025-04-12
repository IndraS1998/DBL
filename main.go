package main

import (
	"fmt"
	"raft/rpc_server"
	"raft/state"
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

	go rpc_server.StartRPCServerListener(n1, &wg)
	go rpc_server.StartRPCServerListener(n2, &wg)
	go rpc_server.StartRPCServerListener(n3, &wg)

	//start various timers
	n1.StartTimer(&wg)
	n2.StartTimer(&wg)
	n3.StartTimer(&wg)

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
			case <-n2.StartElectionChan:
				state.BegginElection(n2)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-n2.StartElectionChan:
				state.BegginElection(n3)
			}
		}
	}()

	// Wait to become leader and erform leader duties
	go func() {
		for {
			fmt.Println("restarting heartbeat sending process! for: ", n1.Address)
			select {
			case <-n1.BecomeLeaderChan:
				fmt.Printf("%s about to send rpcs, \n", n1.Address)
				state.SendHeartbeat(n1)
			case <-n1.RevertToFollowerChan:
				fmt.Printf("stopping timer for %v, revering to follower \n", n1.Address)
				n1.ResetTimerChan <- true
				return
			}
		}
	}()
	go func() {
		for {
			fmt.Println("restarting heartbeat sending process! for: ", n2.Address)
			select {
			case <-n2.BecomeLeaderChan:
				fmt.Printf("%s about to send rpcs, \n", n2.Address)
				state.SendHeartbeat(n2)
			case <-n2.RevertToFollowerChan:
				fmt.Printf("stopping timer for %v, revering to follower \n", n2.Address)
				n2.ResetTimerChan <- true
				return
			}
		}
	}()

	go func() {
		for {
			fmt.Println("restarting heartbeat sending process! for: ", n3.Address)
			select {
			case <-n3.BecomeLeaderChan:
				fmt.Printf("%s about to send rpcs, \n", n3.Address)
				state.SendHeartbeat(n3)
			case <-n3.RevertToFollowerChan:
				fmt.Printf("stopping timer for %v, revering to follower \n", n3.Address)
				n3.ResetTimerChan <- true
				return
			}
		}
	}()
	wg.Wait()
}
