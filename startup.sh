#!/bin/bash

# clean
make clean

# Create data directories
mkdir data/
mkdir data/n1/
mkdir data/n2/
mkdir data/n3/

# Spin up Raft cluster of 3 nodes on ports 8080, 8081, 8082 for testing.
./bin/golemdb -id=1 -kv-addr=localhost:8080 -raft-addr=localhost:4040 -data-dir="data/n1/" -bootstrap=true &
# Sleep for 5s to allow leader some time to Bootstrap cluster before joining.
sleep 5
./bin/golemdb -id=2 -kv-addr=localhost:8081 -raft-addr=localhost:4041 -join-addr=localhost:8080 -data-dir="data/n2/" -bootstrap=false &
./bin/golemdb -id=3 -kv-addr=localhost:8082 -raft-addr=localhost:4042 -join-addr=localhost:8080 -data-dir="data/n3/" -bootstrap=false &

# How to end the processes:
# run `ps` from same shell to get active process or
# `lsof -i tcp:8080`

# and `kill -9 <PID>`