# GolemDB

GolemDB is a distributed key-value database. GolemDB only exposes simple key value APIs like get and put. State and key-value data is persisted locally in [BoltDB](https://github.com/etcd-io/bbolt), and consensus is powered by the [Raft algorithm](https://raft.github.io/), using the [Hashicorp Raft library](https://github.com/hashicorp/raft). This project is built with Golang, and is currently a work in progress. Below is a todo list to track my progress.

A lot of the architecture for this personal project takes inspiration from the [TiKV project](https://github.com/tikv/tikv). As you will see I appropriate some names such as Placement Driver from TiKV and [Google Spanner](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/65b514eda12d025585183a641b5a9e096a3c4be5.pdf).

As it turns out, managing multiple, dynamically-split Raft groups is trickier than it sounds (and it already sounds tricky!). As a result, this project is nothing more than a learning exercise in engineering fault-tolerant and highly available and consistent distributed systems.

### Todo
- [X] Bolt DB
- [X] Multi-Raft
- [X] Range based sharding
- [X] Client redirection using routing table
- [ ] Shard data isolation
- [ ] PD auto sharding

## Architecture

### Sharding
- **Sharding strategy**: In this implementation, data is sharded based on key ranges. Each shard (sometimes referred to as region) is a seperate raft group storing keys within a continuous lexicographical range. (e.g. [a, n), [n, z] may be two shards)
- The node is aware of the shards it is replicating or leading, and should redirecting client requests that it cannot serve to the appropriate node
- The shards are a raft group that could have several different nodes participating in them, and the nodes themselves may participate in many disjoint shards

### Placement Driver
- The Placement Driver (PD) is a seperate module that is responsible for coordinating these shards. It does so by instructing nodes to join, leave, or split shards based on a set of rules.
- PD recieves periodic heartbeats from all the nodes within the GolemDB cluster, which contain info on the nodes status and participation in shards. PD uses this information to make decisions for auto-sharding
- PD maintains and updates a routing table for key ranges to their respective raft groups and leader
     - The routing table may look something like this:
```
[a, g) -> {Shard 1, Node 1}
[g, k) -> {Shard 2, Node 1}
[k, z] -> {Shard 3, Node 2}
```

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
