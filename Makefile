build:
	go build -o bin/goatdb

run: build
	./bin/goatdb

test: 
	@go test -v ./...

clean:
	rm *.db

rpc: 
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./rpc/kv.proto