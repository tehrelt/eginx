build-all: build-lb build-server

preq:
	mkdir -p ./bin

build-server: preq
	go build -o ./bin/server ./cmd/server

build-lb: preq
	go build -o ./bin/loadbalancer ./cmd/app


run-server:
	go run ./cmd/server/main.go -port 5000

run-lb:
	go run ./cmd/app/main.go -config ./configs/config.json