// Raft specific endpoints for managing lifecycle of Raft cluster.
// Todo: add hclog and ..WithLogger option to Raft setup
package srv

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/jeevano/golemdb/internal/client"
	"github.com/jeevano/golemdb/pkg/fsm"
	pb "github.com/jeevano/golemdb/proto/gen"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// startup the raft FSM for a region
func (s *Server) StartRaft(r *Region) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", r.raftAddress)
	if err != nil {
		return fmt.Errorf("Failed to resolve TCP Addr: %v", err)
	}

	// Create the transport
	r.raftTransport, err = raft.NewTCPTransport(
		r.raftAddress,
		tcpAddr,
		s.conf.MaxPool,
		10*time.Second,
		os.Stdout,
	)
	if err != nil {
		return fmt.Errorf("Failed to create Raft transport: %v", err)
	}

	// Create the snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(s.conf.DataDir+"shard_"+strconv.Itoa(int(r.shardId))+"/", 1, os.Stdout)
	if err != nil {
		return fmt.Errorf("Failed to create Raft snapshot store: %v", err)
	}

	// Create the bolt store
	r.raftStore, err = raftboltdb.NewBoltStore(filepath.Join(s.conf.DataDir+"shard_"+strconv.Itoa(int(r.shardId))+"/", "raft.db"))
	if err != nil {
		return fmt.Errorf("Failed to create Raft BoltDB store: %v", err)
	}

	// Create the log cache
	logStore, err := raft.NewLogCache(256, r.raftStore)
	if err != nil {
		return fmt.Errorf("Failed to create Raft log store: %v", err)
	}

	// Set the config
	conf := raft.DefaultConfig()
	conf.LocalID = raft.ServerID(s.conf.ServerId)

	// Create the FSM itself and start Raft
	r.fsm = fsm.NewFSMDatabase(s.db)
	r.raft, err = raft.NewRaft(conf, r.fsm, logStore, r.raftStore, snapshotStore, r.raftTransport)
	if err != nil {
		return fmt.Errorf("Failed to create new Raft: %v", err)
	}

	return nil
}

// Bootstrap a new raft cluster (become leader)
func (s *Server) BootstrapCluster(r *Region) error {
	// Ensure no other peers exist in Raft
	peers, err := r.numVoters()
	if err != nil {
		log.Fatalf("Failed to obtain Raft configuration to initiate Bootstrap: %v", err)
		return fmt.Errorf("Failed to obtain config to Bootstrap cluster: %v", err)
	}
	if peers > 1 {
		log.Fatalf("Already %d peers, cluster already started", peers)
		return fmt.Errorf("Already %d peers. Cluster may have been started already!", peers)
	}

	// Bootstrap the cluster with the current Node
	err = r.raft.BootstrapCluster(raft.Configuration{Servers: []raft.Server{{
		ID:      raft.ServerID(s.conf.ServerId),
		Address: r.raftTransport.LocalAddr(),
	}}}).Error()
	if err != nil {
		log.Fatalf("Failed to complete Raft Bootstrap Cluster: %v", err)
		return fmt.Errorf("Failed to complete Raft Bootstrap Cluster: %v", err)
	}

	log.Printf("Successfully Bootstrapped cluster!")

	return nil
}

// Join the requesting node to the current raft cluster corresponding to requested region/shard
func (s *Server) Join(_ context.Context, req *pb.JoinRequest) (*emptypb.Empty, error) {
	log.Printf("Initiating Join request for Node %s with address %s", req.ServerId, req.Address)

	region := s.regions[req.ShardId]

	// Obtain configuration
	confFuture := region.raft.GetConfiguration()
	if err := confFuture.Error(); err != nil {
		log.Fatalf("Failed to obtain Raft configuration to initiate Join request: %v", err)
		return empty(), fmt.Errorf("Failed to get Raft configuration for Join request: %v", err)
	}

	// Make sure the server does not already exist in the cluster
	for _, srv := range confFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(req.ServerId) && srv.Address == raft.ServerAddress(req.Address) {
			log.Printf("Node %s with address %s has already joined the cluster", req.ServerId, req.Address)
			return empty(), nil
		}

		if srv.ID == raft.ServerID(req.ServerId) || srv.Address == raft.ServerAddress(req.Address) {
			log.Fatalf("A node with a duplicate ID (%s) or Address (%s) already exists in the cluster!", req.ServerId, req.Address)
			return empty(), fmt.Errorf("Failed to join Raft cluster: duplicate address or ID")
		}
	}

	// Add the voter to Raft cluster
	if err := region.raft.AddVoter(raft.ServerID(req.ServerId), raft.ServerAddress(req.Address), 0, 0).Error(); err != nil {
		log.Fatalf("Node %s with address %s failed to Join the cluster: %v", req.ServerId, req.Address, err)
		return empty(), fmt.Errorf("Failed to join Raft: %v", err)
	}

	log.Printf("Succesfully joined Node %s with address %s to Raft cluster!", req.ServerId, req.Address)
	return empty(), nil
}

// Remove the requesting node from the current raft cluster
func (s *Server) Leave(_ context.Context, req *pb.LeaveRequest) (*emptypb.Empty, error) {
	return empty(), fmt.Errorf("Not yet implemented!")
}

// Return status to the requesting raft node
func (s *Server) Status(_ context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	return &pb.StatusResponse{}, fmt.Errorf("Not yet implemented!")
}

// Instruct node to join a shard at another node
func (s *Server) JoinShard(_ context.Context, req *pb.JoinShardRequest) (*emptypb.Empty, error) {
	return empty(), s.JoinShardInternal(req.LeaderAddress, req.Start, req.End, req.ShardId)
}

// Instruct node to split a shard in half (mid point is determined by the node)
func (s *Server) SplitShard(_ context.Context, req *pb.SplitShardRequest) (*emptypb.Empty, error) {
	_, err := s.SplitShardInternal(req.ShardId)
	return empty(), err
}

// Becomes the leader of a new shard
func (s *Server) BootstrapShardInternal(start, end string, shardId int32) error {
	region := s.NewRegion(start, end, true, shardId)

	if err := s.StartRaft(region); err != nil {
		return fmt.Errorf("Failed to start up the Raft FSM: %v", err)
	}

	if err := s.BootstrapCluster(region); err != nil {
		return fmt.Errorf("Failed to Bootstrap cluser: %v", err)
	}

	return nil
}

// Joins a specified shard at the leader
func (s *Server) JoinShardInternal(leaderAddr, start, end string, shardId int32) error {
	region := s.NewRegion(start, end, false, shardId)
	if err := s.StartRaft(region); err != nil {
		return fmt.Errorf("Failed to start up the Raft FSM: %v", err)
	}

	client, close, err := client.NewRaftClient(leaderAddr)
	if err != nil {
		return fmt.Errorf("Failed to create raft client: %v", err)
	}

	err = client.Join(s.conf.ServerId, region.raftAddress, shardId)

	if err != nil {
		close()
		return fmt.Errorf("Failed to join Raft cluster: %v", err)
	}

	close()

	return nil
}

// Splits everything after the split point into a new shard, and returns shardId
// Returns error if not the leader or on failure
func (s *Server) SplitShardInternal(shardId int32) (int32, error) {
	return -1, fmt.Errorf("Not yet implemented!")
}

func (r *Region) numVoters() (int, error) {
	configFuture := r.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return 0, fmt.Errorf("Failed to get Raft config: %v", err)
	}
	// Count the voters
	var n int
	for _, srv := range configFuture.Configuration().Servers {
		if srv.Suffrage == raft.Voter {
			n++
		}
	}
	return n, nil
}

func (r *Region) shardLeader() (bool, error) {
	// todo: how can I actually tell if I am leader
	return r.isLeader, nil
}

func empty() *emptypb.Empty {
	return &emptypb.Empty{}
}
