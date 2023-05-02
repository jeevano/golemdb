package srv

import (
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/jeevano/golemdb/pkg/config"
	"github.com/jeevano/golemdb/pkg/db"
	"github.com/jeevano/golemdb/pkg/fsm"
	"github.com/jeevano/golemdb/pkg/pd"
	pb "github.com/jeevano/golemdb/proto/gen"
	"log"
	"sync"
	"time"
)

type Server struct {
	pb.KvServer
	pb.RaftServer
	db             *db.Database      // DB for storing KV inputs
	conf           *config.Config    // General Raft cluster config
	pdClient       *pd.PDClient      // Placement driver client to heartbeat to PD
	regions        map[int32]*Region // The regions (shards) this node is a member of
	availRaftAddrs map[string]bool   // set of unused raft addresses for creating new regions
	mu             sync.Mutex
}

type Region struct {
	start         string
	end           string
	isLeader      bool
	shardId       int32
	raftAddress   string
	raft          *raft.Raft             // Raft
	raftTransport *raft.NetworkTransport // Network transport for Node communication
	raftStore     *raftboltdb.BoltStore  // Bolt store for Raft logs
	fsm           *fsm.FSMDatabase       // Raft FSM for consensus over DB
}

func NewServer(db *db.Database, conf *config.Config) *Server {
	s := &Server{
		db:             db,
		conf:           conf,
		regions:        make(map[int32]*Region),
		availRaftAddrs: make(map[string]bool),
	}

	for _, a := range conf.RaftAddresses {
		s.availRaftAddrs[a] = true
	}

	return s
}

func (s *Server) NewRegion(start, end string, isLeader bool, shardId int32) *Region {
	// Pick an unused address for the region
	var addr string = ""
	for k, _ := range s.availRaftAddrs {
		addr = k
		delete(s.availRaftAddrs, k)
		break
	}
	if addr == "" {
		return nil
	}

	// create the region
	s.regions[shardId] = &Region{
		start:       start,
		end:         end,
		isLeader:    isLeader,
		shardId:     shardId,
		raftAddress: addr,
	}

	return s.regions[shardId]
}

func (s *Server) Start() error {
	// Initiate heartbeat Goroutine
	go s.heartbeatRoutine()

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
