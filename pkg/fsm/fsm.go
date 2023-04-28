// FSM for committing raft log entries to local Bolt DB
// Used in conjuction with kv_server to maintain consensus / data replication of keys among nodes
package fsm

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"github.com/jeevano/golemdb/pkg/db"
	"io"
	"log"
)

type FSMDatabase struct {
	db *db.Database
}

type Op string

const (
	PutOp Op = "put"
	GetOp Op = "get"
)

type Event struct {
	Op  Op     `json:"op,omitempty"`
	Key []byte `json:"key,omitempty"`
	Val []byte `json:"val,omitempty"`
}

// Invoked once a log entry is committed
func (fsm *FSMDatabase) Apply(logEntry *raft.Log) interface{} {
	var e Event
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		log.Fatalf("Failed to unmarshal raft log entry: %v", err)
		return nil
	}

	switch e.Op {
	case PutOp:
		fsm.db.Put(e.Key, e.Val)
		break
	case GetOp:
		// nothing to do?
		break
	default:
		break
	}

	return nil
}

// Nothing to do for snapshot since persisted to disk via BoltDB
func (fsm *FSMDatabase) Snapshot() (raft.FSMSnapshot, error) {
	return newSnapshot()
}

// Used to restore an FSM from snapshot
func (fsm *FSMDatabase) Restore(rClose io.ReadCloser) error {
	defer rClose.Close()
	// Not yet implemented...
	return nil
}

func NewFSMDatabase(db *db.Database) *FSMDatabase {
	return &FSMDatabase{db: db}
}
