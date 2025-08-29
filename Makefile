# Makefile for arkadia-mapsnap
# Builds a static mapsnap binary in the project root

.PHONY: all clean build

# Allow overriding via environment: e.g. GOOS=linux GOARCH=amd64 make
GOOS ?=
GOARCH ?=
CGO_ENABLED ?= 0

# Output binary name in project root
BINARY := mapsnap

# Version string (optional); falls back to git describe if available
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Common build flags
BUILD_FLAGS := -trimpath
LDFLAGS := -s -w -X main.version=$(VERSION)
# Force static linking (no dynamic libs) and disable cgo via environment

all: $(BINARY)

build: $(BINARY)

$(BINARY):
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) \
		go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS) -extldflags "-static"' \
		-o $(BINARY) ./cmd/mapsnap

clean:
	rm -f $(BINARY)
