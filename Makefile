db:
	go build -o bin/golemdb ./cmd/golemdb

ctl:
	go build -o bin/golemctl ./cmd/golemctl

pd:
	go build -o bin/golempd ./cmd/golempd

all: db pd ctl

run: all
	./dev/startup.sh

test: 
	@go test -v ./...

clean:
	rm -f *.db
	rm -rf data/

rpc: 
	protoc -I=proto --go_out=. --go-grpc_out=. proto/*.proto