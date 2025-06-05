package main

import (
	"fmt"
	"raft/api_server"
	"raft/rpc_server"
	"raft/state"
	"sync"
	"time"
)

func main() {
	peers := []string{"9001", "9002", "9003"}
	wg := sync.WaitGroup{}
	wg.Add(3)
	// create nodes
	n1, err1 := state.NewNode("9001", peers)
	if err1 != nil {
		panic(err1)
	}
	n1APIServer := api_server.NewApiServer(n1)
	if apiErr := n1APIServer.Run(":8001"); apiErr != nil {
		panic(apiErr)
	}

	n1.PrintDetails()

	n2, err2 := state.NewNode("9002", peers)
	if err2 != nil {
		panic(err2)
	}
	n2APIServer := api_server.NewApiServer(n2)
	if apiErr := n2APIServer.Run(":8002"); apiErr != nil {
		panic(apiErr)
	}
	n2.PrintDetails()

	n3, err3 := state.NewNode("9003", peers)
	if err3 != nil {
		panic(err3)
	}
	n3APIServer := api_server.NewApiServer(n3)
	if apiErr := n3APIServer.Run(":8003"); apiErr != nil {
		panic(apiErr)
	}
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
				n1.BeginElection()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-n2.StartElectionChan:
				n2.BeginElection()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-n3.StartElectionChan:
				n3.BeginElection()
			}
		}
	}()

	// Wait to become leader and erform leader duties
	go func() {
		for {
			select {
			case <-n1.BecomeLeaderChan:
				fmt.Printf("%s became leader. Starting heartbeat loop.\n", n1.Address)

			heartbeatLoop:
				for {
					select {
					default:
						time.Sleep(300 * time.Millisecond)
						n1.AppendEntry()
					case <-n1.RevertToFollowerChan:
						fmt.Printf("Reverting %v to follower\n", n1.Address)
						n1.ResetTimerChan <- true
						break heartbeatLoop
					}
				}

			case <-n1.RevertToFollowerChan:
				// fallback just in case Revert is received when not leader
				n1.ResetTimerChan <- true
			}
		}
	}()
	go func() {
		for {
			select {
			case <-n2.BecomeLeaderChan:
				fmt.Printf("%s became leader. Starting heartbeat loop.\n", n2.Address)
			heartbeatLoop:
				for {
					select {
					default:
						time.Sleep(300 * time.Millisecond)
						n2.AppendEntry()

					case <-n2.RevertToFollowerChan:
						fmt.Printf("Reverting %v to follower\n", n2.Address)
						n2.ResetTimerChan <- true
						break heartbeatLoop
					}
				}

			case <-n2.RevertToFollowerChan:
				// fallback just in case Revert is received when not leader
				n2.ResetTimerChan <- true
			}
		}
	}()

	go func() {
		for {
			select {
			case <-n3.BecomeLeaderChan:
				fmt.Printf("%s became leader. Starting heartbeat loop.\n", n3.Address)

			heartbeatLoop:
				for {
					select {
					default:
						time.Sleep(300 * time.Millisecond)
						n3.AppendEntry()
					case <-n3.RevertToFollowerChan:
						fmt.Printf("Reverting %v to follower\n", n3.Address)
						n3.ResetTimerChan <- true
						break heartbeatLoop
					}
				}

			case <-n3.RevertToFollowerChan:
				// fallback just in case Revert is received when not leader
				n3.ResetTimerChan <- true
			}
		}
	}()
	wg.Wait()
}
