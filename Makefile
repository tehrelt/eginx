build-all: build-lb build-backend build-api

preq:
	mkdir -p ./bin

build-backend: preq
	go build -o ./bin/backend ./cmd/backend

build-lb: preq
	go build -o ./bin/eginx ./cmd/app

build-api: preq
	go build -o ./bin/api ./cmd/api

run-backend:
	go run ./cmd/backend/main.go -port 5000

run-lb:
	go run ./cmd/app/main.go -config ./configs/config.json

run-api:
	go run ./cmd/api/main.go -config ./configs/config.json

clean:
	rm -rf ./bin