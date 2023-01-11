.DEFAULT_GOAL := gen
.PHONY: run test lint gen breaking wc

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

breaking: # detect breaking proto changes
	@buf breaking --against ".git#branch=main"

wc:
	@find . -name '*.go' -not -name '*_test.go' -not -name "*.connect.go" -not -name "*.pb.go" -not -name "*.pb.validate.go" | xargs wc
