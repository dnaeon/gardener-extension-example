.DEFAULT_GOAL := build

# Set SHELL to bash and configure options
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

GOCMD?= go
SRC_ROOT := $(shell git rev-parse --show-toplevel)
HACK_DIR := $(SRC_ROOT)/hack
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

# ENVTEST_K8S_VERSION configures the version of Kubernetes, which will be
# installed by setup-envtest.
#
# In order to configure the Kubernetes version to match the version used by the
# k8s.io/api package, use the following setting.
#
# ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{ printf "1.%d.%d", $$3, $$4 }')
#
# Or set the version here explicitly.
ENVTEST_K8S_VERSION ?= 1.34.1

# Common options for the `addlicense' tool
ADDLICENSE_OPTS ?= -f $(HACK_DIR)/LICENSE_BOILERPLATE.txt \
			-ignore "dev/**" \
			-ignore "**/*.md" \
			-ignore "**/*.html" \
			-ignore "**/*.yaml" \
			-ignore "**/*.yml" \
			-ignore "**/Dockerfile"

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
	@echo "Setting up envtest for Kubernetes version v$(ENVTEST_K8S_VERSION) ..."
	@KUBEBUILDER_ASSETS="$$( $(GO_TOOL) setup-envtest use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCAL_BIN) -p path )" \
		$(GOCMD) test -v -race -coverprofile=coverage.txt -covermode=atomic $(shell $(GOCMD) list ./pkg/...)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE):$(IMAGE_TAG) -t $(IMAGE):latest .

.PHONY: update-tools
update-tools:
	$(GOCMD) get -u -modfile $(TOOLS_MOD_FILE) tool

.PHONY: addlicense
addlicense:
	@$(GO_TOOL) addlicense $(ADDLICENSE_OPTS) .

.PHONY: checklicense
checklicense:
	@files=$$( $(GO_TOOL) addlicense -check $(ADDLICENSE_OPTS) .) || { \
		echo "Missing license headers in the following files:"; \
		echo "$${files}"; \
		echo "Run 'make addlicense' in order to fix them."; \
		exit 1; \
	}

.PHONY: generate-operator-extension
generate-operator-extension:
	$(GO_TOOL) extension-generator \
		--name example \
		--component-category extension \
		--provider-type example \
		--destination $(SRC_ROOT)/examples/kustomize/extension/base/extension.yaml \
		--extension-oci-repository $(IMAGE):$(IMAGE_TAG)
	$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/kustomize/extension > $(SRC_ROOT)/examples/operator-extension.yaml

.PHONY: check-helm
check-helm:
	@$(GO_TOOL) helm lint $(SRC_ROOT)/charts
	@$(GO_TOOL) helm template $(SRC_ROOT)/charts | \
		$(GO_TOOL) kubeconform \
			-strict \
			-verbose \
			-summary \
			-output pretty \
			-schema-location default \
			-schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'

.PHONY: check-examples
check-examples:
	@$(GO_TOOL) kubeconform \
		-skip Kustomization \
		-strict \
		-verbose \
		-summary \
		-output pretty \
		-schema-location default \
		-schema-location "$(SRC_ROOT)/test/schemas/{{.Group}}/{{.ResourceAPIVersion}}/{{.ResourceKind}}.json" \
		./examples
