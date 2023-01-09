.DEFAULT_GOAL := gen
.PHONY: test lint gen

test:
	@go test -race -covermode=atomic -coverprofile=coverage.out ./...

lint:
	@if golint ./... 2>&1 | grep '^'; then exit 1; fi; # Requires comments for exported functions
	@golangci-lint run
	@buf lint

gen:
	@buf generate
