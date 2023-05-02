// Entry point of Placement Driver, which is the scheduler of the GolemDB cluster
// The Placement driver joins nodes to each other and splits shards
package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
	placementDriver "github.com/jeevano/golemdb/pkg/pd"
	pb "github.com/jeevano/golemdb/proto/gen"
)

var (
	serverAddr = flag.String("server-addr", "localhost:5555", "The address of the PD server")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *serverAddr)

	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *serverAddr, err)
		return
	}

	pd := placementDriver.NewPlacementDriver()
	s := placementDriver.NewPDServer(pd)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPDServer(grpcServer, s)

	pd.Start()

	grpcServer.Serve(lis)
}
