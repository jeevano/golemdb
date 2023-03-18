// Implements simple gRPC client to make calls to remote database
// TODO: add TLS support
package client

import (
	"context"
	pb "github.com/jeevano/goatdb/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

type KvClient struct {
	client pb.KvClient // private client struct
}

func NewKvClient(serverAddr string) (*KvClient, func() error, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		defer conn.Close()
		return nil, nil, err
	}

	client := pb.NewKvClient(conn)

	return &KvClient{client}, conn.Close, nil
}

func (c *KvClient) Get(key string) (val string, err error) {
	req := pb.GetRequest{Key: []byte(key)}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.client.Get(ctx, &req)
	if err != nil {
		log.Fatalf("fail to get: %v", err)
		return "", err
	}

	return string(res.Val), nil
}

func (c *KvClient) Put(key string, val string) error {
	req := pb.PutRequest{Key: []byte(key), Val: []byte(val)}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.client.Put(ctx, &req)
	if err != nil {
		log.Fatalf("fail to put: %v", err)
	}

	return err
}
