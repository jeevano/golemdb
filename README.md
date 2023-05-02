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

## Placement Driver
- Recieves heartbeats from all the nodes within the GolemDB cluster, containing info on what shards they are members of
- Maintains a routing table for key ranges and their respective raft groups and leader
- Will join nodes (on their behalf) to shard regions, and split regions when they get too large

## Sharding
- **Sharding strategy**: In this implementation, data is sharded based on key ranges
- The Placement Driver is responsible for resharding and coordination of nodes participating in each shard
- The node is aware of the shards it is replicating or leading, and is capable of redirecting client requests to the appropriate node by using the routing table supplied by the Placement Driver
- The shards have a one to many relationship with the GolemDB nodes, and a given node may be a follower or leader of numerous disjoint shards
```
+------------------------------------------+
|                  CLIENT                  |
+------------------------------------------+
     V                V                V      gRPC
+--------+       +--------+       +--------+
| NODE   |       | NODE   |       | NODE   |
|        |       |        |       |        |       +--------+
|--------|       |--------|       |--------|       |        |
| RAFT 1 | <---> | RAFT 1 |       |        |       |   PD   |
|        |       |        |       |        |       |        |
| RAFT 2 | <---> | RAFT 2 | <---> | RAFT 2 |       +--------+
|        |       |        |       |        |
|        |       | RAFT 3 | <---> | RAFT 3 |
+--------+       +--------+       +--------+
```