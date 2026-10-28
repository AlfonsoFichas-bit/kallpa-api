build:
	cd kallpa-api && go build -o bin/api ./cmd/api

run:
	cd kallpa-api && go run ./cmd/api

test:
	cd kallpa-api && go test ./...

test-race:
	cd kallpa-api && go test -race ./...

test-coverage:
	cd kallpa-api && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

lint:
	cd kallpa-api && golangci-lint run

fmt:
	cd kallpa-api && go fmt ./...

.PHONY: build run test test-race test-coverage lint fmt
