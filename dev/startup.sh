#!/bin/bash

# clean
make clean

# Create data directories
mkdir data/
mkdir data/n1/
mkdir data/n2/
mkdir data/n3/

# Spin up Raft cluster of 3 nodes on ports 8080, 8081, 8082 for testing.
./bin/golemdb -id=1 -kv-addr=localhost:8080 -raft-addr-list=localhost:4040,localhost:4041,localhost:4042 -data-dir="data/n1/" -bootstrap=true &
# Sleep for 5s to allow leader some time to Bootstrap cluster before joining.
sleep 5
./bin/golemdb -id=2 -kv-addr=localhost:8081 -raft-addr-list=localhost:4050,localhost:4051,localhost:4052 -join-addr=localhost:8080 -data-dir="data/n2/" -bootstrap=false &
./bin/golemdb -id=3 -kv-addr=localhost:8082 -raft-addr-list=localhost:4060,localhost:4061,localhost:4062 -join-addr=localhost:8080 -data-dir="data/n3/" -bootstrap=false &