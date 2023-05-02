package pd

import (
	"context"
	"encoding/json"
	"fmt"
	pb "github.com/jeevano/golemdb/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type PDClient struct {
	client pb.PDClient
	routingTable *RoutingTable
}

type ShardInfo = pb.ShardInfo

func NewPDClient(pdAddr string) (*PDClient, func() error, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(pdAddr, opts...)
	if err != nil {
		defer conn.Close()
		return nil, nil, fmt.Errorf("fail to dial: %v", err)
	}

	client := pb.NewPDClient(conn)

	return &PDClient{client: client}, conn.Close, nil
}

func (c *PDClient) DoHeartbeat(serverId string, address string, shards []*ShardInfo) error {
	req := pb.HeartbeatRequest{
		ServerId: serverId,
		Address:  address,
		Shards:   shards,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send the heartbeat and retrieve the routing table
	resp, err := c.client.Heartbeat(ctx, &req)
	if err != nil {
		return fmt.Errorf("Failed to do heartbeat: %v", err)
	}

	// Deserialize the routing table
	err = json.Unmarshal(resp.RoutingTable, c.routingTable)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal routing table: %v", err)
	}

	return nil
}
