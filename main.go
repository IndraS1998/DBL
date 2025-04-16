package main

import (
	"fmt"
	"raft/rpc_server"
	"raft/state"
	"sync"
	"time"
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
			select {
			case <-n1.BecomeLeaderChan:
				fmt.Printf("%s became leader. Starting heartbeat loop.\n", n1.Address)

			heartbeatLoop:
				for {
					select {
					default:

						time.Sleep(5 * time.Second)
						ok, err := state.SendHeartbeat(n1)
						if err != nil {
							fmt.Printf("Error sending heartbeat: %v\n", err)
						}
						if !ok {
							fmt.Printf("%s failed to maintain majority, stepping down.\n", n1.Address)
							break heartbeatLoop
						}
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
				fmt.Printf("%s became leader. Starting heartbeat loop.\n", n1.Address)
			heartbeatLoop:
				for {
					select {
					default:
						time.Sleep(5 * time.Second)
						ok, err := state.SendHeartbeat(n2)
						if err != nil {
							fmt.Printf("Error sending heartbeat: %v\n", err)
						}
						if !ok {
							fmt.Printf("%s failed to maintain majority, stepping down.\n", n2.Address)
							break heartbeatLoop
						}
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
						time.Sleep(5 * time.Second)
						ok, err := state.SendHeartbeat(n3)
						if err != nil {
							fmt.Printf("Error sending heartbeat: %v\n", err)
						}
						if !ok {
							fmt.Printf("%s failed to maintain majority, stepping down.\n", n3.Address)
							break heartbeatLoop
						}
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
