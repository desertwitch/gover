# Makefile

BINARY = gover
SRC_DIR = ./cmd/gover

.PHONY: all clean info mocks lint

all: $(BINARY)

$(BINARY):
	CGO_ENABLED=0 GOFLAGS="-mod=vendor" go build -ldflags="-w -s -buildid=" -trimpath -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

info:
	@ldd $(BINARY) || true
	@file $(BINARY)

mocks:
	@find . -type d -name "mocks" -exec rm -vrf {} +
	@mockery --config .mockery.yaml

lint:
	@golangci-lint run

clean:
	@rm -vf $(BINARY) || true
