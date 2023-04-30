// Implements simple gRPC server to allow Get and Put for Key Value pairs in database
// TODO: add tls support
package srv

import (
	"context"
	"encoding/json"
	"github.com/jeevano/golemdb/pkg/fsm"
	pb "github.com/jeevano/golemdb/proto/gen"
	"log"
	"time"
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
	// err := s.db.Put(req.Key, req.Val)

	// Apply the Raft log for DB write to achieve consensus among following Raft nodes
	b, err := json.Marshal(fsm.Event{
		Op:  fsm.PutOp,
		Key: req.Key,
		Val: req.Val,
	})
	if err != nil {
		log.Fatalf("Failed to marshal event: %v", err)
		return new(pb.Empty), err
	}

	if err := s.raft.Apply(b, 10*time.Second).Error(); err != nil {
		log.Fatalf("Failed to apply Raft log: %v", err)
		return new(pb.Empty), err
	}

	return new(pb.Empty), nil
}
