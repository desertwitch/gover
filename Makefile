# Makefile

BINARY = gover
SRC_DIR = ./cmd/gover

.PHONY: all clean info mocks generate-mocks

all: $(BINARY)
mocks: generate-mocks

$(BINARY):
	CGO_ENABLED=0 GOFLAGS="-mod=vendor" go build -ldflags="-w -s -buildid=" -trimpath -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

info:
	@ldd $(BINARY) || true
	@file $(BINARY)

generate-mocks:
	@mockery --config .mockery.yaml

clean:
	@rm -f $(BINARY) || true
