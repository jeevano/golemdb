// Config for each raft node
package config

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	ServerId      string   // Unique ID for the server
	JoinAddress   string   // The address of another raft shard to join on startup
	DataDir       string   // Directory to store Raft and DB files
	KvAddress     string   // Address for KvServer connections
	Bootstrap     bool     // True if this is the first Node to initiate shard
	PdAddress     string   // The address to contact Placement driver
	InitShardId   int32    // The id of the initial shard. All other shardId determined at runtime
	MaxShards     int      // maximum number of shards to participate in (<= 3)
	MaxPool       int      // maximum nodes in a raft cluster
	RaftAddresses []string // reserved addresses to bind raft for each shard
}

func (conf *Config) Load() error {
	flag.StringVar(&conf.ServerId, "id", "", "A unique id for this Raft node.")
	flag.StringVar(&conf.DataDir, "data-dir", "", "Path of where to store FSM state.")
	flag.StringVar(&conf.JoinAddress, "join-addr", "", "Address of another node to send a join request.")
	flag.StringVar(&conf.KvAddress, "kv-addr", "localhost:4444", "Address on which to bind KvServer.")
	flag.StringVar(&conf.PdAddress, "pd-addr", "localhost:5555", "Address on which to contact the Placement Driver.")
	flag.IntVar(&conf.MaxShards, "max-shards", 3, "Max number of shards can join.")
	flag.IntVar(&conf.MaxPool, "max-pool", 7, "Max number of nodes in raft cluster.")
	flag.BoolVar(&conf.Bootstrap, "bootstrap", true, "Bootstrap the cluster with this node.")

	var tmpInitShardId int
	flag.IntVar(&tmpInitShardId, "init-shard-id", 0, "Id of the initial shard")

	var tmpRaftAddresses string
	flag.StringVar(&tmpRaftAddresses, "raft-addr-list", "", "Comma seperated list of reserved raft addresses")

	flag.Parse()

	conf.InitShardId = int32(tmpInitShardId)
	conf.RaftAddresses = strings.Split(tmpRaftAddresses, ",")

	return conf.Validate()
}

func (conf *Config) Validate() error {
	if conf.ServerId == "" {
		return fmt.Errorf("Missing Server Id")
	}
	if conf.JoinAddress == "" && conf.Bootstrap == false {
		return fmt.Errorf("Must supply Join Address if not Bootstrapping cluster")
	}
	if conf.KvAddress == "" {
		return fmt.Errorf("Missing Kv Address")
	}
	if conf.MaxPool > 7 {
		return fmt.Errorf("Max pool is too high!")
	}
	if conf.MaxShards > 3 {
		return fmt.Errorf("Max participation is too high!")
	}
	if len(conf.RaftAddresses) != conf.MaxShards {
		return fmt.Errorf("Mismatch in max participation and reserved addresses")
	}
	return nil
}
