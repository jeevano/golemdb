package pd

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jeevano/golemdb/internal/client"
)

// Placement Driver
type PD struct {
	shards       map[int32]*Shard // map of ShardId to Shard
	nodes        map[string]*Node // map serverid to node info
	routingTable *RoutingTable    // the routing table
	mu           sync.Mutex
}

// GolemDB Node metadata
type Node struct {
	shards       map[int32]*Shard // participating shards
	serverId     string           // identifier of node
	address      string           // raft address of node
	lastHeatbeat int64            // time of last heartbeat for tracking liveliness
}

// Shard metadata
type Shard struct {
	Start      string `json:"start,omitempty"`      // the start of the shard range based on lexicographical-ordering
	End        string `json:"end,omitempty"`        // the end of the shard range based on lexicographical-ordering
	ShardId    int32  `json:"shardId,omitempty"`    // the internal identifier of the shard, assigned by PD
	LeaderAddr string `json:"leaderAddr,omitempty"` // the leader of the shards address (for redirects)
	Size       int32  `json:"size,omitempty"`       // the current size of the shard
	Reads      int32  `json:"reads,omitempty"`      // the number of reads on this shard
	Writes     int32  `json:"writes,omitempty"`     // the number of writes on this shard
}

// Routing table for redirects, JSON fields for byte serialization and de-serialization
type RoutingTable struct {
	Shards []Shard `json:"shards,omitempty"` // list of shards
	valid  bool    // flag for if the cache is invalidated
}

// Initialize the Placement Driver
func NewPlacementDriver() *PD {
	return &PD{
		shards:       make(map[int32]*Shard),
		nodes:        make(map[string]*Node),
		routingTable: &RoutingTable{},
	}
}

func (pd *PD) Start() {
	go pd.reshardRoutine()
}

func (pd *PD) reshardRoutine() {
	for {
		time.Sleep(15 * time.Second)
		pd.Reshard()
	}
}

// Returns a copy of the current routing table
func (pd *PD) getRoutingTable() (RoutingTable, error) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	cpy := make([]Shard, len(pd.routingTable.Shards))
	copy(cpy, pd.routingTable.Shards)

	return RoutingTable{Shards: cpy}, nil
}

// Takes in key as argument, returns leader node address
// Exported, can be used by nodes or client having a cached routing table
func (rt *RoutingTable) Lookup(key string) (string, error) {
	for _, shard := range rt.Shards {
		if strings.Compare(key, shard.Start) >= 0 && strings.Compare(key, shard.End) <= 0 {
			return shard.LeaderAddr, nil
		}
	}

	// Happens if the key starts with an unsupported character or something went wrong
	return "", fmt.Errorf("Could not find suitable shard")
}

// Update routing table based on Node and Shard info from Heartbeats
func (pd *PD) UpdateRoutingTable() error {
	// If the routing table has not been invalidated by a split or leadership change, nothing to do
	if pd.routingTable.valid {
		return nil
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()
	// Check one more time
	if pd.routingTable.valid {
		return nil
	}

	// Delete the current routing table
	pd.routingTable = &RoutingTable{}

	// Copy relevant info the the routing table
	for _, shard := range pd.shards {
		shard_cpy := Shard{
			Start:      shard.Start,
			End:        shard.End,
			LeaderAddr: shard.LeaderAddr,
		}
		pd.routingTable.Shards = append(pd.routingTable.Shards, shard_cpy)
	}

	return nil
}

// Perform consistency check on routing table and nodes
// Forgets nodes that have not heartbeated in last X seconds
// Creates shards wherever needed, and joins nodes to shards as needed
func (pd *PD) Reshard() error {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	// Map of shards to participating nodes
	var shardToNodes = make(map[int32][]*Node)
	var candidatesToJoin []*Node

	// Check if nodes are dead
	for key, n := range pd.nodes {
		if n.lastHeatbeat < time.Now().Unix()-60 {
			delete(pd.nodes, key)
		}
		for key, _ := range n.shards {
			shardToNodes[key] = append(shardToNodes[key], n)
		}

		if len(n.shards) < 2 {
			candidatesToJoin = append(candidatesToJoin, n)
		}
	}

	// Check if Shards need new nodes
	for shardId, nodes := range shardToNodes {
		if len(nodes) < 3 {
			log.Printf("Found shard with less than 3 participating nodes")

			// If less than 3 participating nodes, join a candidate node to the shard
			leaderAddr := pd.shards[shardId].LeaderAddr
			var candidateAddr string = ""
			for _, a := range candidatesToJoin {
				if a.address != leaderAddr && a.shards[shardId] == nil {
					candidateAddr = a.address
					break
				}
			}
			if candidateAddr == "" {
				continue
			}

			// Dial the candidate node
			client, close, err := client.NewRaftClient(candidateAddr)
			if err != nil {
				return fmt.Errorf("Failed to dial node %s: %v", leaderAddr, err)
			}

			log.Printf("Joining node %s to region lead by %s", candidateAddr, leaderAddr)

			// Tell the candidate to join the leader's shard
			// Will cause lock contention: TODO: put this RPC outside of critical section
			err = client.JoinShard(leaderAddr, pd.shards[shardId].Start, pd.shards[shardId].End, shardId)

			if err != nil {
				return fmt.Errorf("Failed to join node %s to cluster with leader %s: %v",
					candidateAddr, leaderAddr, err)
			}

			close()
		}
	}

	// Check if Shards need to be split
	// Not yet implemented

	return nil
}

// Heartbeat handler for nodes to periodically give information to PD
func (pd *PD) HandleHeartbeat(serverId string, address string, shards []*ShardInfo) error {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	// Consistency check on node, and update liveliness time
	node := pd.nodes[serverId]
	if node == nil {
		pd.nodes[serverId] = &Node{
			shards:       make(map[int32]*Shard),
			serverId:     serverId,
			address:      address,
			lastHeatbeat: time.Now().Unix(),
		}
		node = pd.nodes[serverId]
	} else {
		pd.nodes[serverId].lastHeatbeat = time.Now().Unix()
	}

	// Consistency check on shards (if any)
	for _, shard := range shards {
		oldShard := pd.shards[shard.ShardId]
		// Check if the shard has been encountered before
		// If the leader heartbeats, then add the shard to list
		// What if leader crashed? -> Raft should elect new leader
		if oldShard == nil && !shard.IsLeader {
			continue
		} else if oldShard == nil {
			pd.shards[shard.ShardId] = &Shard{
				Start:      shard.Start,
				End:        shard.End,
				ShardId:    shard.ShardId,
				LeaderAddr: address,
				Size:       shard.Size,
				Reads:      shard.Reads,
				Writes:     shard.Writes,
			}
			oldShard = pd.shards[shard.ShardId]
		}

		// update the shard if leader
		if shard.IsLeader {
			oldShard.Size = shard.Size
			oldShard.Reads = shard.Reads
			oldShard.Writes = shard.Writes
		}

		// The shard is known in PD but not consistent with node, update the node
		if oldShard != nil && node.shards[shard.ShardId] == nil {
			node.shards[shard.ShardId] = oldShard
		}
	}

	// Invalidate the routing table cache
	pd.routingTable.valid = false

	return nil
}
