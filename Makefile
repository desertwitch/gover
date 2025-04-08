# Makefile

BINARY = gover
SRC_DIR = ./cmd/gover

VERSION := $(shell git rev-parse --short=7 HEAD)

.PHONY: all clean check debug help info lint mocks test vendor

all: vendor mocks check $(BINARY) ## Runs the entire build chain for the application

$(BINARY): ## Builds the application
	CGO_ENABLED=0 GOFLAGS="-mod=vendor" go build -ldflags="-w -s -X main.Version=$(VERSION) -buildid=" -trimpath -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

check: ## Runs all static analysis and tests on the application code
	@$(MAKE) lint
	@$(MAKE) test

clean: ## Returns the application build stage to its original state (deleting files)
	@rm -vf $(BINARY) || true

clean-mocks: ## Returns the mock build stage to its original state (deleting files)
	@find . -type f -name "mock_*.go" -exec rm -vf {} +
	@find . -type d -name "mocks" -exec rm -vrf {} +

debug: ## Builds the application in debug mode (with symbols, race checks, ...)
	CGO_ENABLED=1 GOFLAGS="-mod=vendor" go build -ldflags="-X main.Version=$(VERSION)-DBG" -trimpath -race -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

help: ## Shows all build related commands of the Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

info: ## Shows information about the application binaries that were built
	@ldd $(BINARY) || true
	@file $(BINARY)

lint: ## Runs the linter on the application code
	@golangci-lint cache clean
	@golangci-lint run

mocks: ## Generates the mocks for the application code
	@mockery --config .mockery.yaml

test: ## Runs all written tests for and on the application code
	@go test -failfast -race -covermode=atomic ./...

test-coverage: ## Runs all coverage tests for and on the application code
	@go test -failfast -race -covermode=atomic -coverprofile=coverage.tmp ./... && \
	grep -v "mock_" coverage.tmp > coverage.txt && \
	rm coverage.tmp

vendor: ## Pulls the (remote) dependencies into the local vendor folder
	@go mod tidy
	@go mod vendor
