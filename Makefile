.DEFAULT_GOAL := gen
.PHONY: run-local-server run-local-bot test lint gen docker breaking wc

run-local-server:
	@go run cmd/server/main.go

run-local-bot:
	@go run cmd/server/main.go -runDiscord -cleanUpDiscordCommands

test:
	@go test -race -covermode=atomic -coverprofile=coverage.out ./...

lint:
	@if golint ./... 2>&1 | grep '^'; then exit 1; fi; # Requires comments for exported functions
	@golangci-lint run
	@buf lint

gen:
	@buf format -w
	@buf generate

docker:
	@docker build --tag bettor:$(shell git rev-parse HEAD | cut -c1-8) .

breaking: # detect breaking proto changes
	@buf breaking --against ".git#branch=main"

wc:
	@find . -name '*.go' -not -name '*_test.go' -not -name "*.connect.go" -not -name "*.pb.go" -not -name "*.pb.validate.go" | xargs wc
