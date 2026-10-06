# Recipes must stay portable: GNU Make runs them with cmd.exe on Windows and sh elsewhere.
# No shell syntax (VAR=x prefixes, &&, rm, mkdir): use target-specific exported variables instead.

DATA ?= ./data

# Explicit Go packages: ./... would walk into web/node_modules, which ships stray .go files.
GOPKGS = ./cmd/... ./internal/... ./web

.PHONY: test lint test-go lint-go fmt-go web web-install lint-web test-web run build build-go build-linux build-windows

## test: run all tests
test: test-go test-web

## lint: run all linters (zero issues tolerated)
lint: lint-go lint-web

test-go:
	go test $(GOPKGS)

lint-go:
	golangci-lint run $(GOPKGS)

## fmt-go: format Go code (gofumpt + goimports)
fmt-go:
	golangci-lint fmt $(GOPKGS)

## web-install: install frontend dependencies
web-install:
	npm --prefix web ci

## web: build the frontend into web/dist (embedded by the Go build)
web:
	npm --prefix web run build

lint-web:
	npm --prefix web run lint

test-web:
	npm --prefix web run test

## run: serve $(DATA) on 127.0.0.1:8080 without opening a browser
run:
	go run ./cmd/open-niscat --data $(DATA) --open-browser=false

GOBUILD = go build -trimpath -ldflags "-s -w"

## build: build the binaries for Linux and Windows (amd64) into dist/
build: web build-go

build-go: build-linux build-windows

build-linux: export GOOS = linux
build-linux: export GOARCH = amd64
build-linux: export CGO_ENABLED = 0
build-linux:
	$(GOBUILD) -o dist/open-niscat-linux-amd64 ./cmd/open-niscat

build-windows: export GOOS = windows
build-windows: export GOARCH = amd64
build-windows: export CGO_ENABLED = 0
build-windows:
	$(GOBUILD) -o dist/open-niscat-windows-amd64.exe ./cmd/open-niscat
