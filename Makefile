build:
	go build -o bin/golemdb ./cmd/golemdb

run: build
	./bin/golemdb

test: 
	@go test -v ./...

clean:
	rm -f *.db
	rm -rf data/

ctl:
	go build -o bin/golemctl ./cmd/golemctl

.PHONY: rpc
rpc: 
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./internal/rpc/kv.proto
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./internal/rpc/raft.proto