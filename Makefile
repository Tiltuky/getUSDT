GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin

.PHONY: build test docker-build run start lint clean

build:
	go build -o app ./cmd

test:
	go test -v ./...

docker-build:
	docker compose build

start:
	docker compose up -d

run:
	docker compose run --rm app ./main

lint:
	golangci-lint run
