.PHONY: proto generate build run

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/gophkeeper.proto

generate: proto

.PHONY: build-client
build-client:
	go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date +'%Y-%m-%d_%H:%M:%S') -X main.buildCommit=$(git rev-parse HEAD)" -o bin/client ./client/cmd/client

.PHONY: build-server
build-server:
	go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date +'%Y-%m-%d_%H:%M:%S') -X main.buildCommit=$(git rev-parse HEAD)" -o bin/server ./server/cmd/server

.PHONY: build
build: build-client build-server

test:
	go test ./...

.PHONY: run-with-db
run-with-db: build-server run-postgresql 
		GOPHKEEPER_DSA=postgres://admin:admin@127.0.0.1:5432/gophkeeper?sslmode=disable \
		bin/server

.PHONY: run-postgresql
run-postgresql: 
	docker rm -f $$(docker ps -q  -f=name=postgres) || true
	docker run -d --name postgres \
	  -e POSTGRES_USER=admin \
	  -e POSTGRES_PASSWORD=admin \
	  -e POSTGRES_DB=gophkeeper \
	  -p 127.0.0.1:5432:5432 \
	  -v postgres-data:/var/lib/postgresql/data \
	  postgres:17
	sleep 5