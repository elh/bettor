.DEFAULT_GOAL := gen
.PHONY: run test lint gen

run:
	@go run cmd/server.go # run from binary

test:
	@go test -race -covermode=atomic -coverprofile=coverage.out ./...

lint:
	@if golint ./... 2>&1 | grep '^'; then exit 1; fi; # Requires comments for exported functions
	@golangci-lint run
	@buf lint

gen:
	@buf format -w
	@buf generate
