# Recipes must stay portable: GNU Make runs them with cmd.exe on Windows and sh elsewhere.
# No shell syntax (VAR=x prefixes, &&, rm, mkdir): use target-specific exported variables instead.

DATA ?= ./data

.PHONY: test lint test-go lint-go fmt-go run

## test: run all tests
test: test-go

## lint: run all linters (zero issues tolerated)
lint: lint-go

test-go:
	go test ./...

lint-go:
	golangci-lint run ./...

## fmt-go: format Go code (gofumpt + goimports)
fmt-go:
	golangci-lint fmt ./...

## run: serve $(DATA) on 127.0.0.1:8080 without opening a browser
run:
	go run ./cmd/open-niscat --data $(DATA) --open-browser=false
