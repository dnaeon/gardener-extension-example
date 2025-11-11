.DEFAULT_GOAL := build

GOCMD?= go
SRC_ROOT := $(shell git rev-parse --show-toplevel)
SRC_DIRS := $(shell $(GOCMD) list -f '{{ .Dir }}' ./...)

GOOS := $(shell $(GOCMD) env GOOS)
GOARCH := $(shell $(GOCMD) env GOARCH)
TOOLS_MOD_DIR := $(SRC_ROOT)/internal/tools
TOOLS_MOD_FILE := $(TOOLS_MOD_DIR)/go.mod
GO_TOOL := $(GOCMD) tool -modfile $(TOOLS_MOD_FILE)

LOCAL_BIN ?= $(SRC_ROOT)/bin
BINARY    ?= $(LOCAL_BIN)/extension

VERSION := $(shell cat VERSION)
EFFECTIVE_VERSION ?= $(VERSION)-$(shell git rev-parse --short HEAD)
ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

IMAGE     ?= europe-docker.pkg.dev/gardener-project/public/gardener/extensions/example
IMAGE_TAG ?= $(EFFECTIVE_VERSION)

$(LOCAL_BIN):
	mkdir -p $(LOCAL_BIN)

.PHONY: goimports-reviser
goimports-reviser:
	$(GO_TOOL) goimports-reviser -set-exit-status -rm-unused ./...

.PHONY: lint
lint:
	$(GO_TOOL) golangci-lint run --config=$(SRC_ROOT)/.golangci.yaml ./...

$(BINARY): $(SRC_DIRS) | $(LOCAL_BIN)
	$(GOCMD) build \
		-o $(LOCAL_BIN)/ \
		-ldflags="-X 'gardener-extension-example/pkg/version.Version=${EFFECTIVE_VERSION}'" \
		./cmd/extension

.PHONY: build
build: $(BINARY)

.PHONY: run
run: $(BINARY)
	$(BINARY) manager

.PHONY: get
get:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

.PHONY: test
test:
	$(GOCMD) test -v -race ./...

.PHONY: test-cover
test-cover:
	$(GOCMD) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE):$(IMAGE_TAG) -t $(IMAGE):latest .

.PHONY: update-tools
update-tools:
	$(GOCMD) get -u -modfile $(TOOLS_MOD_FILE) tool
