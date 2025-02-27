# Makefile

BINARY = gover
SRC_DIR = ./cmd/gover

.PHONY: all clean info

all: $(BINARY)

$(BINARY):
	CGO_ENABLED=0 GOFLAGS="-mod=vendor" go build -ldflags="-w -s -buildid=" -trimpath -o $(BINARY) $(SRC_DIR)
	@$(MAKE) info

info:
	@ldd $(BINARY) || true
	@file $(BINARY)

clean:
	@rm -f $(BINARY) || true
