.PHONY: fmt vet test build run

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

build:
	go build ./...

run:
	go run ./cmd/server
