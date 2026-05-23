.PHONY: lint test build

lint:
	golangci-lint run ./...

test:
	go test ./...

build:
	go build ./...

