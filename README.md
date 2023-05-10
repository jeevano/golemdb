# GolemDB

GolemDB is a distributed key-value database. GolemDB only exposes simple key value APIs like get and put. State and key-value data is persisted locally in [BoltDB](https://github.com/etcd-io/bbolt), and consensus is powered by the [Raft algorithm](https://raft.github.io/), using the [Hashicorp Raft library](https://github.com/hashicorp/raft). This project is built with Golang, and is currently a work in progress. Below is a todo list to track my progress.

A lot of the architecture for this personal project takes inspiration from the [TiKV project](https://github.com/tikv/tikv). As you will see I appropriate some names such as Placement Driver from TiKV and [Google Spanner](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/65b514eda12d025585183a641b5a9e096a3c4be5.pdf).

As it turns out, managing multiple, dynamically-split Raft groups is trickier than it sounds (and it already sounds tricky!). As a result, this project is nothing more than a learning exercise in engineering fault-tolerant and highly available and consistent distributed systems.

### Todo
- [X] Bolt DB and gRPC key-value server
- [X] Multi-Raft management for consensus over multiple shards
- [X] Range based sharding by keys
- [X] Placement Driver (PD) cluster management 
- [X] Routing table and client redirection 
- [ ] PD auto-sharding and resharding 

## Architecture

### Sharding
- **Sharding strategy**: In this implementation, data is sharded based on key ranges. Each shard (sometimes referred to as region) is a seperate raft group storing keys within a continuous lexicographical range. (e.g. [a, n), [n, z] may be two shards)
- The GolemDB server node keeps track of the shards it is leading or replicating, and will serve client requests that fall in these shard ranges, and reject all other requests
- The shards are a raft group that could have several nodes participating in them, and the nodes themselves may participate in many different shards.

### Placement Driver
- The Placement Driver (PD) is a seperate module that is responsible for coordinating these shards. It does so by instructing nodes to join, leave, or split shards based on a set of rules.
- PD recieves periodic heartbeats from all the nodes within the GolemDB cluster, which contain info on the nodes status and participation in shards. PD uses this information to perform auto-sharding and resharding.
- PD maintains and updates a routing table for key ranges to their respective raft groups and leader.
     - The routing table may look something like this:
```
[a, g) -> {Shard 1, Node 1}
[g, k) -> {Shard 2, Node 1}
[k, z] -> {Shard 3, Node 2}
```

### Client
- The client is provided an initial node to contact for the first user request. If the requested key is in a shard which the GolemDB node does not have access to, the node will respond with the up-to-date routing table of shards and their leaders
- The client uses the cached routing table to request the correct node for all subsequent user GET and PUT operations. If a node cannot serve the request as a result of re-sharding, again it will respond with an up-to-date routing table for the client to cache and retry.

### Overview
Below is a top-down view of a GolemDB cluster.
```
+------------------------------------------+
|                  CLIENT                  |
+------------------------------------------+
     V                V                V      gRPC
+--------+       +--------+       +--------+
| NODE 1 |       | NODE 2 |       | NODE 3 |
|        |       |        |       |        |       +--------+
|--------|       |--------|       |--------|       |        |
| RAFT 1 | <---> | RAFT 1 |       |        | }---> |   PD   |
|        |       |        |       |        |       |        |
| RAFT 2 | <---> | RAFT 2 | <---> | RAFT 2 |       +--------+
|        |       |        |       |        |
|        |       | RAFT 3 | <---> | RAFT 3 |
+--------+       +--------+       +--------+
```
