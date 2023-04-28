# GolemDB

## Todo
- [X] Bolt
- [X] gRPC
- [X] Raft consensus
- [ ] Sharding
- [ ] Containerize

## Nice to have
- [ ] MVCC
- [ ] Better observability

## Sharding
- **Top Half**: All nodes members of same raft cluster, for having consensus over configuration and sharding
  - Join shard, leave shard, reshard, etc.
- **Bottom Half**: Raft clusters for maintaining consensus over the actual data within shards. A subset of GolemDB nodes may be a member of these clusters.
  - Put operations, delete operations, get operations.
- Each node maintains a dynamic global look-up-table for caching the location where keys should be read/stored based on the sharding strategy. This table is synchronized on the top-half Raft FSM of each node, and used for redirecting user requests to Raft nodes that can serve them.
- **Sharding strategy**: In this implementation, data is sharded based on key ranges. The nodes will additionally perform resharding when shards grow too large. 
```
+--------+       +--------+       +--------+
| RAFT   | <---> | RAFT   | <---> | RAFT   |
|        |       |        |       |        |
|--------|       |--------|       |--------|
| RAFT 1 | <---> | RAFT 1 |       |        |
|        |       |        |       |        |
| RAFT 2 | <---> | RAFT 2 | <---> | RAFT 2 |
|        |       |        |       |        |
|        |       | RAFT 3 | <---> | RAFT 3 |
+--------+       +--------+       +--------+
```