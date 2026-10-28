build:
	cd kallpa-api && go build -o bin/api ./cmd/api

run:
	cd kallpa-api && go run ./cmd/api

test:
	cd kallpa-api && go test ./...

lint:
	cd kallpa-api && golangci-lint run

fmt:
	cd kallpa-api && go fmt ./...

.PHONY: build run test lint fmt
