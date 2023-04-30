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
	protoc -I=proto --go_out=. --go-grpc_out=. proto/*.proto