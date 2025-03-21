# Makefile

BINARY = gover
SRC_DIR = ./cmd/gover

.PHONY: all clean info mocks lint debug

all: $(BINARY)

$(BINARY):
	@$(MAKE) mocks
	CGO_ENABLED=0 GOFLAGS="-mod=vendor" go build -ldflags="-w -s -buildid=" -trimpath -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

debug:
	@$(MAKE) mocks
	CGO_ENABLED=1 GOFLAGS="-mod=vendor" go build -trimpath -race -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

info:
	@ldd $(BINARY) || true
	@file $(BINARY)

mocks:
	@mockery --config .mockery.yaml

lint:
	@golangci-lint run

clean:
	@find . -type d -name "mocks" -exec rm -vrf {} +
	@rm -vf $(BINARY) || true
