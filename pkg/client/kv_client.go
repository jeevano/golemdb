// Implements simple gRPC client to make calls to remote database
// TODO: add TLS support
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeevano/golemdb/pkg/pd"
	pb "github.com/jeevano/golemdb/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KvClient struct {
	initAddr     string
	routingTable *pd.RoutingTable // cached routing table for key regions
}

func NewKvClient(serverAddr string) *KvClient {
	return &KvClient{
		initAddr:     serverAddr,
		routingTable: &pd.RoutingTable{},
	}
}

func (c *KvClient) Get(key string) (val string, err error) {
	addr, err := c.routingTable.Lookup(key)
	if err != nil || addr == "" {
		addr = c.initAddr
	}
	res, err := doGet(key, c.initAddr)
	if err != nil {
		return "", err
	}
	if res.Success {
		return string(res.Val), nil
	}
	// If still no result, then retry
	if err = json.Unmarshal(res.RoutingTable, c.routingTable); err != nil {
		return "", err
	}
	addr, err = c.routingTable.Lookup(key)
	if err != nil {
		return "", err
	}
	res, err = doGet(key, addr)
	if err != nil {
		return "", err
	}
	return string(res.Val), nil
}

func doGet(key string, addr string) (*pb.GetResponse, error) {
	// Dial the server
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(addr, opts...)
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("fail to dial: %v", err)
	}

	client := pb.NewKvClient(conn)

	// Get request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.GetRequest{Key: []byte(key)}
	res, err := client.Get(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *KvClient) Put(key string, val string) error {
	addr, err := c.routingTable.Lookup(key)
	if err != nil || addr == "" {
		addr = c.initAddr
	}
	res, err := doPut(key, val, c.initAddr)
	if err != nil {
		return err
	}
	if res.Success {
		return nil
	}
	// If still no result, then retry
	if err = json.Unmarshal(res.RoutingTable, c.routingTable); err != nil {
		return err
	}
	addr, err = c.routingTable.Lookup(key)
	if err != nil {
		return err
	}
	res, err = doPut(key, val, addr)
	if err != nil {
		return err
	}
	return nil
}

func doPut(key, val, addr string) (*pb.PutResponse, error) {
	// Dial the server
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(addr, opts...)
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("fail to dial: %v", err)
	}

	client := pb.NewKvClient(conn)

	// Put request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.PutRequest{Key: []byte(key), Val: []byte(val)}

	res, err := client.Put(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
