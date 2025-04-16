package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// --- Raft Message Types ---
type AppendEntriesArgs struct {
	Term int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

// --- MockRPCClient ---
type MockRPCClient struct {
	mu          sync.RWMutex
	partitioned map[string]bool
	peers       map[string]*RaftNode
}

func NewMockRPCClient() *MockRPCClient {
	return &MockRPCClient{
		partitioned: make(map[string]bool),
		peers:       make(map[string]*RaftNode),
	}
}

func (m *MockRPCClient) RegisterPeer(id string, node *RaftNode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[id] = node
}

func (m *MockRPCClient) SetPartitioned(peerID string, status bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.partitioned[peerID] = status
}

func (m *MockRPCClient) SendHeartbeat(fromID, toID string, args *AppendEntriesArgs) (*AppendEntriesReply, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.partitioned[toID] {
		return nil, fmt.Errorf("network unreachable: partitioned")
	}
	peer, ok := m.peers[toID]
	if !ok {
		return nil, fmt.Errorf("peer not found")
	}
	return peer.ReceiveHeartbeat(args)
}

// --- FakeTimer ---
type FakeTimer struct {
	timeout  time.Duration
	elapsed  time.Duration
	callback func()
	mu       sync.Mutex
}

func NewFakeTimer(timeout time.Duration, onTimeout func()) *FakeTimer {
	return &FakeTimer{
		timeout:  timeout,
		callback: onTimeout,
	}
}

func (ft *FakeTimer) Advance(d time.Duration) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.elapsed += d
	if ft.elapsed >= ft.timeout && ft.callback != nil {
		ft.callback()
		ft.elapsed = 0
	}
}

// --- RaftNode ---
type RaftNode struct {
	ID    string
	state string
	rpc   *MockRPCClient
	mu    sync.Mutex
	term  int
}

func NewRaftNode(id string) *RaftNode {
	return &RaftNode{ID: id, state: "follower"}
}

func (rn *RaftNode) SetRPCClient(rpc *MockRPCClient) {
	rn.rpc = rpc
}

func (rn *RaftNode) StartElection() {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.state = "candidate"
	fmt.Printf("[%s] Starting election\n", rn.ID)
}

func (rn *RaftNode) BecomeLeader() {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.state = "leader"
	fmt.Printf("[%s] Became leader\n", rn.ID)
}

func (rn *RaftNode) ReceiveHeartbeat(args *AppendEntriesArgs) (*AppendEntriesReply, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	if args.Term >= rn.term {
		rn.state = "follower"
		rn.term = args.Term
	}
	return &AppendEntriesReply{Term: rn.term, Success: true}, nil
}

func (rn *RaftNode) IsLeader() bool    { return rn.state == "leader" }
func (rn *RaftNode) IsCandidate() bool { return rn.state == "candidate" }
func (rn *RaftNode) IsFollower() bool  { return rn.state == "follower" }

func (rn *RaftNode) SendHeartbeatTo(toID string) {
	args := &AppendEntriesArgs{Term: rn.term}
	_, _ = rn.rpc.SendHeartbeat(rn.ID, toID, args)
}

// --- TestCluster ---
type TestCluster struct {
	nodes           map[string]*RaftNode
	rpcClient       *MockRPCClient
	timers          map[string]*FakeTimer
	electionTimeout time.Duration
}

func NewTestCluster(nodeIDs []string, timeout time.Duration) *TestCluster {
	rpc := NewMockRPCClient()
	cluster := &TestCluster{
		nodes:           make(map[string]*RaftNode),
		timers:          make(map[string]*FakeTimer),
		rpcClient:       rpc,
		electionTimeout: timeout,
	}

	for _, id := range nodeIDs {
		node := NewRaftNode(id)
		node.SetRPCClient(rpc)
		rpc.RegisterPeer(id, node)
		cluster.nodes[id] = node
		timer := NewFakeTimer(timeout, func() {
			node.StartElection()
		})
		cluster.timers[id] = timer
	}

	return cluster
}

func (tc *TestCluster) PartitionNode(id string) {
	tc.rpcClient.SetPartitioned(id, true)
}

func (tc *TestCluster) HealPartition(id string) {
	tc.rpcClient.SetPartitioned(id, false)
}

func (tc *TestCluster) AdvanceTime(nodeID string, d time.Duration) {
	tc.timers[nodeID].Advance(d)
}

func (tc *TestCluster) ElectLeader(id string) *RaftNode {
	node := tc.nodes[id]
	node.BecomeLeader()
	return node
}

func (tc *TestCluster) GetNode(id string) *RaftNode {
	return tc.nodes[id]
}

// --- Actual Unit Test ---
func Test_TemporaryNetworkPartition(t *testing.T) {
	cluster := NewTestCluster([]string{"Leader", "Follower1", "Follower2"}, 150*time.Millisecond)
	leader := cluster.ElectLeader("Leader")

	cluster.PartitionNode("Follower1")

	t.Run("Short Partition", func(t *testing.T) {
		cluster.AdvanceTime("Follower1", 100*time.Millisecond)
		require.True(t, leader.IsLeader())
		require.False(t, cluster.GetNode("Follower1").IsCandidate())
	})

	t.Run("Long Partition", func(t *testing.T) {
		cluster.AdvanceTime("Follower1", 300*time.Millisecond)
		t.Logf("Follower1 state: %s", cluster.GetNode("Follower1").state)
		//require.True(t, cluster.GetNode("Follower1").IsFollower())
		require.True(t, cluster.GetNode("Follower1").IsCandidate())
	})

	t.Run("Post Recovery", func(t *testing.T) {
		cluster.HealPartition("Follower1")
		leader.SendHeartbeatTo("Follower1")
		require.True(t, cluster.GetNode("Follower1").IsFollower())
	})
}
