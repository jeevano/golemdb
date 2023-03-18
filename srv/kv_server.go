// Implements simple gRPC server to allow Get and Put for Key Value pairs in database
// TODO: add tls support
package srv

import (
	"context"
	"github.com/jeevano/golemdb/db"
	pb "github.com/jeevano/golemdb/rpc"
	"log"
)

type KvServer struct {
	pb.KvServer
	db *db.Database // the backing database
}

func NewKvServer(db *db.Database) *KvServer {
	return &KvServer{db: db}
}

func (s *KvServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	log.Printf("Recieved get request of key %v", string(req.Key))
	val, err := s.db.Get(req.Key)
	if err != nil {
		log.Fatalf("Failed to serve get request: %v", err)
		return new(pb.GetResponse), err
	}
	return &pb.GetResponse{Val: val}, nil
}

func (s *KvServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.Empty, error) {
	log.Printf("Recieved put request of kv pair {%v : %v}", string(req.Key), string(req.Val))
	err := s.db.Put(req.Key, req.Val)
	if err != nil {
		log.Fatalf("Failed to serve put request: %v", err)
	}
	return new(pb.Empty), err
}
