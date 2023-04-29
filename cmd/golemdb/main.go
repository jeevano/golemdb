// Entry point of the golemdb service - initializes database and server
package main

import (
	"github.com/jeevano/golemdb/pkg/config"
	"github.com/jeevano/golemdb/pkg/db"
	"github.com/jeevano/golemdb/pkg/client"
	"github.com/jeevano/golemdb/pkg/srv"
	pb "github.com/jeevano/golemdb/internal/rpc"
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

	// Fire up the raft FSM
	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start up the Raft FSM: %v", err)
		return
	}

	if conf.Bootstrap == true {
		// Bootstrap the Raft cluster with this node
		if err := s.BootstrapCluster(); err != nil {
			log.Fatalf("Failed to Bootstrap cluser: %v", err)
			return
		}
		log.Printf("Successfully bootstrapped Raft cluster!")
	} else {
		// Join the Raft cluster with the supplied Join addresss
		client, close, err := client.NewRaftClient(conf.JoinAddress)
		if err != nil {
			log.Fatalf("Failed to create raft client: %v", err)
		}

		err = client.Join(conf.ServerId, conf.RaftAddress)

		if err != nil {
			log.Fatalf("Failed to join Raft cluster: %v", err)
			close()
			return
		}
		
		close()
		log.Printf("Successfully joined cluster at peer %s!", conf.JoinAddress)
	}

	// And begin serving incoming Raft requests and Kv requests
	log.Printf("Kv Server listening on %s", conf.KvAddress)
	log.Printf("Raft listening on %s", conf.RaftAddress)
	grpcServer.Serve(lis)
}
