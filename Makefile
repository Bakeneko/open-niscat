# Recipes must stay portable: GNU Make runs them with cmd.exe on Windows and sh elsewhere.
# No shell syntax (VAR=x prefixes, &&, rm, mkdir): use target-specific exported variables instead.

DATA ?= ./data
PYTHON ?= python

# Explicit Go packages: ./... would walk into web/node_modules, which ships stray .go files.
GOPKGS = ./cmd/... ./internal/... ./web

.PHONY: test lint vuln test-go lint-go fmt-go web web-install lint-web test-web tools-install lint-tools test-tools run build build-go build-linux-amd64 build-linux-arm64 build-windows-amd64 build-darwin-arm64 docker

## test: run all tests
test: test-go test-web test-tools

## lint: run all linters (zero issues tolerated)
lint: lint-go lint-web lint-tools

test-go:
	go test $(GOPKGS)

lint-go:
	golangci-lint run $(GOPKGS)

## vuln: report known vulnerabilities reachable from the Go code (needs network access)
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 $(GOPKGS)

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

# Data tooling (tools/data): Python 3.12+; install its dependencies once with make tools-install.
## tools-install: install the data tooling and its checkers (pillow, ruff, mypy, pytest)
tools-install:
	$(PYTHON) -m pip install -e tools/data[dev]

lint-tools:
	$(PYTHON) -m ruff check tools/data
	$(PYTHON) -m ruff format --check tools/data
	$(PYTHON) -m mypy --config-file tools/data/pyproject.toml tools/data

test-tools:
	$(PYTHON) -m pytest -q tools/data

## run: serve $(DATA) on 127.0.0.1:8080 without opening a browser
run:
	go run ./cmd/open-niscat --data $(DATA) --open-browser=false

# Version shown by --version, the startup log and the home page: the git tag, else the commit (-dirty when
# modified). Override with make build VERSION=v1.2.3.
VERSION ?= $(shell git describe --tags --always --dirty)
GOBUILD = go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)"

## build: build the release binaries (Linux amd64/arm64, Windows amd64, macOS arm64) into dist/
build: web build-go

build-go: build-linux-amd64 build-linux-arm64 build-windows-amd64 build-darwin-arm64

build-linux-amd64: export GOOS = linux
build-linux-amd64: export GOARCH = amd64
build-linux-amd64: export CGO_ENABLED = 0
build-linux-amd64:
	$(GOBUILD) -o dist/open-niscat-linux-amd64 ./cmd/open-niscat

build-linux-arm64: export GOOS = linux
build-linux-arm64: export GOARCH = arm64
build-linux-arm64: export CGO_ENABLED = 0
build-linux-arm64:
	$(GOBUILD) -o dist/open-niscat-linux-arm64 ./cmd/open-niscat

build-windows-amd64: export GOOS = windows
build-windows-amd64: export GOARCH = amd64
build-windows-amd64: export CGO_ENABLED = 0
build-windows-amd64:
	$(GOBUILD) -o dist/open-niscat-windows-amd64.exe ./cmd/open-niscat

build-darwin-arm64: export GOOS = darwin
build-darwin-arm64: export GOARCH = arm64
build-darwin-arm64: export CGO_ENABLED = 0
build-darwin-arm64:
	$(GOBUILD) -o dist/open-niscat-darwin-arm64 ./cmd/open-niscat

## docker: build the Docker image for this machine's platform, tagged open-niscat:$(VERSION)
docker:
	docker build --build-arg VERSION=$(VERSION) -t open-niscat:$(VERSION) .
