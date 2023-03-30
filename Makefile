build:
	go build -o bin/golemdb

run: build
	./bin/golemdb

test: 
	@go test -v ./...

clean:
	rm *.db

.PHONY: ctl
ctl:
	go build -o bin/golemctl ./ctl/golemctl.go 

.PHONY: rpc
rpc: 
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./rpc/kv.proto
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./rpc/raft.proto