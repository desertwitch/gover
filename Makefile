# Makefile

BINARY = gover
SRC_DIR = ./cmd/gover

VERSION := $(shell git rev-parse --short=7 HEAD)

.PHONY: all clean check debug help info lint mocks test vendor

all: check mocks $(BINARY)

$(BINARY):
	CGO_ENABLED=0 GOFLAGS="-mod=vendor" go build -ldflags="-w -s -X=main.Version=$(VERSION) -buildid=" -trimpath -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

check:
	@$(MAKE) lint
	@$(MAKE) test

clean:
	@find . -type d -name "mocks" -exec rm -vrf {} +
	@rm -vf $(BINARY) || true

debug:
	CGO_ENABLED=1 GOFLAGS="-mod=vendor" go build -ldflags="-X=main.Version=$(VERSION)-DBG" -trimpath -race -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

info:
	@ldd $(BINARY) || true
	@file $(BINARY)

lint:
	@golangci-lint run

mocks:
	@mockery --config .mockery.yaml

test:
	@go test -race ./...

vendor:
	@go mod tidy
	@go mod vendor
