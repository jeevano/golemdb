// Raft specific endpoints for managing lifecycle of Raft cluster.
// Todo: add hclog and ..WithLogger option to Raft setup
package srv

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	pb "github.com/jeevano/golemdb/rpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"log"
)

// Bootstrap a new raft cluster (become leader)
func (s *Server) BootstrapCluster() error {
	// Ensure no other peers exist in Raft
	peers, err := s.numVoters()
	if err != nil {
		log.Fatalf("Failed to obtain Raft configuration to initiate Bootstrap: %v", err)
		return fmt.Errorf("Failed to obtain config to Bootstrap cluster: %v", err)
	}
	if peers > 1 {
		log.Fatalf("Already %d peers, cluster already started", peers)
		return fmt.Errorf("Already %d peers. Cluster may have been started already!", peers)
	}

	// Bootstrap the cluster with the current Node
	err = s.raft.BootstrapCluster(raft.Configuration{Servers: []raft.Server{{
		ID:      raft.ServerID(s.conf.ServerId),
		Address: s.raftTransport.LocalAddr(),
	}}}).Error()
	if err != nil {
		log.Fatalf("Failed to complete Raft Bootstrap Cluster: %v", err)
		return fmt.Errorf("Failed to complete Raft Bootstrap Cluster: %v", err)
	}

	log.Printf("Successfully Bootstrapped cluster!")

	return nil
}

// Join the requesting node to the current raft cluster
func (s *Server) Join(_ context.Context, req *pb.JoinRequest) (*emptypb.Empty, error) {
	log.Printf("Initiating Join request for Node %s with address %s", req.ServerId, req.Address)

	// Obtain configuration
	confFuture := s.raft.GetConfiguration()
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
	if err := s.raft.AddVoter(raft.ServerID(req.ServerId), raft.ServerAddress(req.Address), 0, 0).Error(); err != nil {
		log.Fatalf("Node %s with address %s failed to Join the cluster: %v", req.ServerId, req.Address, err)
		return empty(), fmt.Errorf("Failed to join Raft: %v", err)
	}

	log.Printf("Succesfully joined Node %s with address %s to Raft cluster!", req.ServerId, req.Address)
	return empty(), nil
}

// Remove the requesting node from the current raft cluster
func (s *Server) Remove(_ context.Context, req *pb.RemoveRequest) (*emptypb.Empty, error) {
	return empty(), fmt.Errorf("Not yet implemented!")
}

// Return status to the requesting raft node
func (s *Server) Status(_ context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	return &pb.StatusResponse{}, fmt.Errorf("Not yet implemented!")
}

func (s *Server) numVoters() (int, error) {
	configFuture := s.raft.GetConfiguration()
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

func empty() *emptypb.Empty {
	return &emptypb.Empty{}
}
