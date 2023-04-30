build:
	go build -o bin/golemdb ./cmd/golemdb

ctl:
	go build -o bin/golemctl ./cmd/golemctl

pd:
	go build -o bin/golempd ./cmd/golempd

all: build pd ctl

run: build
	./bin/golemdb

test: 
	@go test -v ./...

clean:
	rm -f *.db
	rm -rf data/

rpc: 
	protoc -I=proto --go_out=. --go-grpc_out=. proto/*.proto