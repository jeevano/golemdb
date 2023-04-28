package srv

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/jeevano/golemdb/pkg/config"
	"github.com/jeevano/golemdb/pkg/db"
	"github.com/jeevano/golemdb/pkg/fsm"
	pb "github.com/jeevano/golemdb/internal/rpc"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Server struct {
	pb.KvServer
	pb.RaftServer
	raft          *raft.Raft             // Raft
	raftTransport *raft.NetworkTransport // Network transport for Node communication
	raftStore     *raftboltdb.BoltStore  // Bolt store for Raft logs
	fsm           *fsm.FSMDatabase       // Raft FSM for consensus over DB
	db            *db.Database           // DB for storing KV inputs
	conf          *config.Config         // General Raft cluster config
}

func NewServer(db *db.Database, conf *config.Config) *Server {
	return &Server{
		db:   db,
		conf: conf,
	}
}

func (s *Server) Start() error {
	// Startup the Raft FSM
	// No startup process needed for KvServer
	if err := s.startRaft(); err != nil {
		log.Fatalf("Failed to startup Raft FSM: %v", err)
		return err
	}
	return nil
}

// startup the raft FSM
func (s *Server) startRaft() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.conf.RaftAddress)
	if err != nil {
		return fmt.Errorf("Failed to resolve TCP Addr: %v", err)
	}

	// Create the transport
	s.raftTransport, err = raft.NewTCPTransport(
		s.conf.RaftAddress,
		tcpAddr,
		s.conf.MaxPool,
		10*time.Second,
		os.Stdout,
	)
	if err != nil {
		return fmt.Errorf("Failed to create Raft transport: %v", err)
	}

	// Create the snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(s.conf.DataDir, 1, os.Stdout)
	if err != nil {
		return fmt.Errorf("Failed to create Raft snapshot store: %v", err)
	}

	// Create the bolt store
	s.raftStore, err = raftboltdb.NewBoltStore(filepath.Join(s.conf.DataDir, "raft.db"))
	if err != nil {
		return fmt.Errorf("Failed to create Raft BoltDB store: %v", err)
	}

	// Create the log cache
	logStore, err := raft.NewLogCache(256, s.raftStore)
	if err != nil {
		return fmt.Errorf("Failed to create Raft log store: %v", err)
	}

	// Set the config
	conf := raft.DefaultConfig()
	conf.LocalID = raft.ServerID(s.conf.ServerId)

	// Create the FSM itself and start Raft
	s.fsm = fsm.NewFSMDatabase(s.db)
	s.raft, err = raft.NewRaft(conf, s.fsm, logStore, s.raftStore, snapshotStore, s.raftTransport)
	if err != nil {
		return fmt.Errorf("Failed to create new Raft: %v", err)
	}

	return nil
}
