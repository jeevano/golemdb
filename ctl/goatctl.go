// Command line tool to get and put into db
/*
Usage:
	./goatctl --server-addr=localhost:4444 --op=put --key=key1 --val=hello!
	./goatctl --server-addr=localhost:4444 --op=get --key=key1
	hello!
*/
package main

import (
	"flag"
	"fmt"
	"github.com/jeevano/goatdb/client"
)

var (
	serverAddr = flag.String("server-addr", "localhost:4444", "The address of the grpc server")
	op         = flag.String("op", "", "The operation performed (get or put)")
	key        = flag.String("key", "", "The key")
	val        = flag.String("val", "", "The value")
)

func main() {
	flag.Parse()

	client, close, _ := client.NewKvClient(*serverAddr)
	defer close()

	if *op == "get" {
		res, _ := client.Get(*key)
		fmt.Printf("%v\n", res)
	} else if *op == "put" {
		client.Put(*key, *val)
	}
}