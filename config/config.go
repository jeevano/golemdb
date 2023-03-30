// Config for each raft node
package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ServerId    string // Unique ID for the server
	RaftAddress string // The address for this node to advertise
	JoinAddress string // The address of another Node to join
	DataDir     string // Directory to store Raft and DB files
	KvAddress   string // Address for KvServer connections
	MaxPool     int    // Max number of Raft node connections
	Bootstrap   bool   // True if this is the first Node to initiate cluster
}

func (conf *Config) Load() error {
	flag.StringVar(&conf.ServerId, "id", "", "A unique id for this Raft node.")
	flag.StringVar(&conf.DataDir, "data-dir", "", "Path of where to store FSM state.")
	flag.StringVar(&conf.JoinAddress, "join-addr", "", "Address of another node to send a join request.")
	flag.StringVar(&conf.RaftAddress, "raft-addr", "", "Raft address of the node.")
	flag.StringVar(&conf.KvAddress, "kv-addr", "localhost:4444", "Address on which to bind KvServer.")
	flag.IntVar(&conf.MaxPool, "max-pool", 7, "Max pool of Raft Nodes.")
	flag.BoolVar(&conf.Bootstrap, "bootstrap", true, "Bootstrap the cluster with this node.")
	flag.Parse()
	return conf.Validate()
}

func (conf *Config) Validate() error {
	if conf.ServerId == "" {
		return fmt.Errorf("Missing Server Id")
	}
	if conf.RaftAddress == "" {
		return fmt.Errorf("Missing Raft Address")
	}
	if conf.JoinAddress == "" && conf.Bootstrap == false {
		return fmt.Errorf("Must supply Join Address if not Bootstrapping cluster")
	}
	if conf.KvAddress == "" {
		return fmt.Errorf("Missing Kv Address")
	}
	if conf.MaxPool < 1 {
		return fmt.Errorf("Maxpool must be at least 1")
	}
	if conf.KvAddress == conf.RaftAddress || conf.KvAddress == conf.JoinAddress || conf.RaftAddress == conf.JoinAddress {
		return fmt.Errorf("Duplicate address")
	}
	return nil
}
