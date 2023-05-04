package raft

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jeevano/golemdb/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RaftClient struct {
	client pb.RaftClient
}

func NewRaftClient(serverAddr string) (*RaftClient, func() error, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		defer conn.Close()
		return nil, nil, fmt.Errorf("fail to dial: %v", err)
	}

	client := pb.NewRaftClient(conn)

	return &RaftClient{client}, conn.Close, nil
}

func (c *RaftClient) Join(serverId, address string, shardId int32) error {
	req := pb.JoinRequest{
		ServerId: serverId,
		Address:  address,
		ShardId:  shardId,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.client.Join(ctx, &req)

	return err
}

func (c *RaftClient) JoinShard(address, start, end string, shardId int32) error {
	req := pb.JoinShardRequest{
		LeaderAddress: address,
		Start:         start,
		End:           end,
		ShardId:       shardId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.client.JoinShard(ctx, &req)

	return err
}

func (c *RaftClient) Leave(serverId, address string, shardId int32) error {
	return fmt.Errorf("Not yet implemented")
}
