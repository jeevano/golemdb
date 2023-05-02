// Entry point of the golemdb service - initializes database and server
package main

import (
	"github.com/jeevano/golemdb/pkg/config"
	"github.com/jeevano/golemdb/pkg/db"
	"github.com/jeevano/golemdb/pkg/srv"
	pb "github.com/jeevano/golemdb/proto/gen"
	"google.golang.org/grpc"
	"log"
	"net"
	"path/filepath"
)

var conf = &config.Config{}

func main() {
	if err := conf.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
		return
	}

	// Create the backing Kv database
	db, close, err := db.NewDatabase(filepath.Join(conf.DataDir, "golem.db"))
	if err != nil {
		log.Fatalf("Failed to create db: %v", err)
		return
	}
	defer close()

	// Startup gRPC server
	lis, err := net.Listen("tcp", conf.KvAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		return
	}

	// Instantiate the server
	var s *srv.Server = srv.NewServer(db, conf)

	// Register the gRPC servers
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterKvServer(grpcServer, s)
	pb.RegisterRaftServer(grpcServer, s)

	log.Print("Starting server...")

	// Fire up the raft FSM (different process if leader or voter)

	if conf.Bootstrap == true {
		// Bootstrap the Raft cluster with this node
		err := s.BootstrapShardInternal("A", "Z", 0)
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Printf("Successfully bootstrapped Raft cluster")
	} else {
		// Join the Raft cluster with the supplied Join addresss
		err := s.JoinShardInternal(conf.JoinAddress, "A", "Z", 0)
		if err != nil {
			log.Fatalf("%v", err)
			return
		}
		log.Printf("Successfully joined cluster at peer %s!", conf.JoinAddress)
	}

	// Start the server (begin heartbeating)
	// s.Start()

	// And begin serving incoming Raft requests and Kv requests
	log.Printf("Kv Server listening on %s", conf.KvAddress)
	grpcServer.Serve(lis)
}
