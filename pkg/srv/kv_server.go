// Implements simple gRPC server to allow Get and Put for Key Value pairs in database
// TODO: add tls support
package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jeevano/golemdb/pkg/fsm"
	pb "github.com/jeevano/golemdb/proto/gen"
	"log"
	"time"
	"strings"
)

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	log.Printf("Recieved get request of key %v", string(req.Key))
	val, err := s.db.Get(req.Key)
	if err != nil {
		log.Fatalf("Failed to serve get request: %v", err)
		return new(pb.GetResponse), err
	}
	return &pb.GetResponse{Val: val}, nil
}

func (s *Server) Put(ctx context.Context, req *pb.PutRequest) (*pb.Empty, error) {
	log.Printf("Recieved put request of kv pair {%v : %v}", string(req.Key), string(req.Val))

	// Check if the key fits into my regions
	region, err := s.findRegion(string(req.Key))
	if err != nil {
		log.Printf("Key %v is not in my range", string(req.Key))
		return empty2(), err
	}

	// Apply the Raft log for DB write to achieve consensus among following Raft nodes
	b, err := json.Marshal(fsm.Event{
		Op:  fsm.PutOp,
		Key: req.Key,
		Val: req.Val,
	})
	if err != nil {
		log.Fatalf("Failed to marshal event: %v", err)
		return empty2(), err
	}

	if err := region.raft.Apply(b, 10*time.Second).Error(); err != nil {
		log.Fatalf("Failed to apply Raft log: %v", err)
		return empty2(), err
	}

	return empty2(), nil
}

func (s *Server) findRegion(key string) (*Region, error) {
	var region *Region = nil
	for _, r := range s.regions {
		if strings.Compare(key, r.start) >= 0 && strings.Compare(key, r.end) <= 0 {
			region = r
			break
		}
	}
	// If the region is still not found, check the routing table
	if region == nil {
		return nil, fmt.Errorf("Not in my regions")
	}

	return region, nil
}

func empty2() *pb.Empty {
	return new(pb.Empty)
}
