// Command line tool to get and put into db
/*
Usage:
	./golemctl --server-addr=localhost:4444 --op=put --key=key1 --val=hello!
	./golemctl --server-addr=localhost:4444 --op=get --key=key1
	hello!
*/
package main

import (
	"flag"
	"fmt"

	"github.com/jeevano/golemdb/pkg/client"
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
		res, err := client.Get(*key)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%v\n", res)
		}
	} else if *op == "put" {
		err := client.Put(*key, *val)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}
