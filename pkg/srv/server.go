package srv

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/jeevano/golemdb/pkg/config"
	"github.com/jeevano/golemdb/pkg/db"
	"github.com/jeevano/golemdb/pkg/fsm"
	"github.com/jeevano/golemdb/pkg/pd"
	pb "github.com/jeevano/golemdb/proto/gen"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Server struct {
	pb.KvServer
	pb.RaftServer
	db       *db.Database      // DB for storing KV inputs
	conf     *config.Config    // General Raft cluster config
	pdClient *pd.PDClient      // Placement driver client to heartbeat to PD
	regions  map[int32]*Region // The regions (shards) this node is a member of
}

type Region struct {
	start         string
	end           string
	isLeader      bool
	shardId       int32
	raft          *raft.Raft             // Raft
	raftTransport *raft.NetworkTransport // Network transport for Node communication
	raftStore     *raftboltdb.BoltStore  // Bolt store for Raft logs
	fsm           *fsm.FSMDatabase       // Raft FSM for consensus over DB
}

func NewServer(db *db.Database, conf *config.Config) *Server {
	return &Server{
		db:      db,
		conf:    conf,
		regions: make(map[int32]*Region),
	}
}

func (s *Server) NewRegion(start, end string, isLeader bool, shardId int32) *Region {
	s.regions[shardId] = &Region{
		start:    start,
		end:      end,
		isLeader: isLeader,
		shardId:  shardId,
	}

	return s.regions[shardId]
}

func (s *Server) Start() error {
	// Initiate heartbeat Goroutine
	go s.heartbeatRoutine()

	return nil
}

func (s *Server) StartRaft(r *Region) error {
	// Startup the Raft FSM for a new region
	if err := s.startRaft(r); err != nil {
		log.Fatalf("Failed to startup Raft FSM: %v", err)
		return err
	}

	return nil
}

// Goroutine for sending heartbeat to PD every 20s
func (s *Server) heartbeatRoutine() {
	log.Printf("Initiating heartbeat routing to PD")
	client, close, err := pd.NewPDClient(s.conf.PdAddress)
	if err != nil {
		log.Fatalf("Failed to initiate heartbeat routine, cannot dial PD at %s: %v", s.conf.PdAddress, err)
		return
	}
	s.pdClient = client
	defer close()

	for {
		// Sleep for 20s in between heartbeats
		time.Sleep(20 * time.Second)
		log.Printf("Sending a heartbeat!")

		// Build the ShardInfo list to send over
		info := make([]*pb.ShardInfo, len(s.regions))
		for _, r := range s.regions {
			shardInfo := &pb.ShardInfo{
				ShardId:  r.shardId,
				IsLeader: r.isLeader,
				Start:    r.start,
				End:      r.end,
			}
			info = append(info, shardInfo)
		}

		// Send heartbeat to PD
		if err := client.DoHeartbeat(s.conf.ServerId, s.conf.KvAddress, info); err != nil {
			log.Fatalf("Failed to send heartbeat to PD: %v", err)
		}
	}
}

// startup the raft FSM for a region
func (s *Server) startRaft(r *Region) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.conf.RaftAddress)
	if err != nil {
		return fmt.Errorf("Failed to resolve TCP Addr: %v", err)
	}

	// Create the transport
	r.raftTransport, err = raft.NewTCPTransport(
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
	r.raftStore, err = raftboltdb.NewBoltStore(filepath.Join(s.conf.DataDir, "raft.db"))
	if err != nil {
		return fmt.Errorf("Failed to create Raft BoltDB store: %v", err)
	}

	// Create the log cache
	logStore, err := raft.NewLogCache(256, r.raftStore)
	if err != nil {
		return fmt.Errorf("Failed to create Raft log store: %v", err)
	}

	// Set the config
	conf := raft.DefaultConfig()
	conf.LocalID = raft.ServerID(s.conf.ServerId)

	// Create the FSM itself and start Raft
	r.fsm = fsm.NewFSMDatabase(s.db)
	r.raft, err = raft.NewRaft(conf, r.fsm, logStore, r.raftStore, snapshotStore, r.raftTransport)
	if err != nil {
		return fmt.Errorf("Failed to create new Raft: %v", err)
	}

	return nil
}
