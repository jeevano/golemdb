package pd

import (
	"context"
	"encoding/json"
	pb "github.com/jeevano/golemdb/proto/gen"
	"log"
)

type PDServer struct {
	pb.PDServer
	pd *PD
}

func NewPDServer(placementDriver *PD) *PDServer {
	return &PDServer{
		pd: placementDriver,
	}
}

// Recieves heartbeats from GolemDB Nodes containing information on participation in shards, and other shard metadata
func (s *PDServer) Heartbeat(_ context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	// Handle the heartbeat
	if err := s.pd.HandleHeartbeat(req.ServerId, req.Address, req.Shards); err != nil {
		log.Fatalf("Failed to handle heartbeat for server %s: %v", req.ServerId, err)
		return &pb.HeartbeatResponse{}, err
	}

	// Update the routing table if needed (i.e. leadership change)
	if err := s.pd.UpdateRoutingTable(); err != nil {
		log.Fatalf("Failed to update routing table: %v", err)
		return &pb.HeartbeatResponse{}, err
	}

	// Retrieve the cached routing table
	rt, err := s.pd.getRoutingTable()
	if err != nil {
		log.Fatalf("Failed to retrieve routing table: %v", err)
		return &pb.HeartbeatResponse{}, err
	}

	// Encode the routing table to binary for sending back to client
	encodedRt, err := json.Marshal(rt)
	if err != nil {
		log.Fatalf("Failed to encode routing table: %v", err)
		return &pb.HeartbeatResponse{}, err
	}

	return &pb.HeartbeatResponse{RoutingTable: encodedRt}, nil
}
