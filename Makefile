build-all: build-lb build-server

preq:
	mkdir -p ./bin

build-backend: preq
	go build -o ./bin/backend ./cmd/server

build-lb: preq
	go build -o ./bin/loadbalancer ./cmd/app


run-backend:
	go run ./cmd/backend/main.go -port 5000

run-lb:
	go run ./cmd/app/main.go -config ./configs/config.json