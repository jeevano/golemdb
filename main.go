// Entry point of the goatdb service - initializes database and server
package main

import (
	"flag"
	"github.com/jeevano/goatdb/db"
	pb "github.com/jeevano/goatdb/rpc"
	"github.com/jeevano/goatdb/srv"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	dbPath     = flag.String("db-path", "a.db", "The db location")
	serverAddr = flag.String("server-addr", "localhost:4444", "The server address")
)

func main() {
	flag.Parse()

	// create the database
	db, close, err := db.NewDatabase(*dbPath)
	if err != nil {
		log.Fatalf("Failed to create db: %v", err)
	}
	defer close()

	// startup gRPC server
	lis, err := net.Listen("tcp", *serverAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterKvServer(grpcServer, srv.NewKvServer(db))

	log.Print("Starting server...")

	grpcServer.Serve(lis)
}
