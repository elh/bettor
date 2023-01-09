.DEFAULT_GOAL := gen
.PHONY: test lint gen

test:
	go test ./...

lint:
	golint ./... # I like the exported comment warnings. Doesn't fail.
	golangci-lint run
	buf lint

gen:
	buf generate
