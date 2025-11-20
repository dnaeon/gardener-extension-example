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
GO_MODULE := $(shell $(GOCMD) list -m -f '{{ .Path }}' )
GO_TOOL := $(GOCMD) tool -modfile $(TOOLS_MOD_FILE)

LOCAL_BIN ?= $(SRC_ROOT)/bin
BINARY    ?= $(LOCAL_BIN)/extension

VERSION := $(shell cat VERSION)
REVISION := $(shell git rev-parse --short HEAD)
EFFECTIVE_VERSION := $(VERSION)-$(REVISION)
ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

# Name and tag for the extension image
IMAGE     ?= europe-docker.pkg.dev/gardener-project/public/gardener/extensions/example
IMAGE_TAG ?= $(EFFECTIVE_VERSION)

# Name and version of the Gardener extension.
EXTENSION_NAME ?= gardener-extension-example
EXTENSION_VERSION ?= $(EFFECTIVE_VERSION)

# Registry used for local development
LOCAL_REGISTRY ?= garden.local.gardener.cloud:5001
# Name of the kind cluster for local development
KIND_CLUSTER ?= gardener-local

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

# Common options for the `kubeconform' tool
KUBECONFORM_OPTS ?= 	-strict \
			-verbose \
			-summary \
			-output pretty \
			-skip Kustomization \
			-schema-location default

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
	@$(GO_TOOL) golangci-lint run --config=$(SRC_ROOT)/.golangci.yaml ./...

$(BINARY): $(SRC_DIRS) | $(LOCAL_BIN)
	$(GOCMD) build \
		-o $(LOCAL_BIN)/ \
		-ldflags="-X '$(GO_MODULE)/pkg/version.Version=${EFFECTIVE_VERSION}'" \
		./cmd/extension

.PHONY: build
build: $(BINARY)

.PHONY: run
run: $(BINARY)
	$(BINARY) manager

.PHONY: get
get:
	@$(GOCMD) mod download
	@$(GOCMD) mod tidy

.PHONY: test
test:
	@echo "Setting up envtest for Kubernetes version v$(ENVTEST_K8S_VERSION) ..."
	@KUBEBUILDER_ASSETS="$$( $(GO_TOOL) setup-envtest use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCAL_BIN) -p path )" \
		$(GOCMD) test -v -race -coverprofile=coverage.txt -covermode=atomic $(shell $(GOCMD) list ./pkg/...)

.PHONY: docker-build
docker-build:
	@docker build \
		--build-arg BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
		--build-arg VERSION=$(VERSION) \
		--build-arg REVISION=$(REVISION) \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):$(IMAGE_TAG) \
		-t $(IMAGE):latest .

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

.PHONY: generate
generate:
	@$(GO_TOOL) controller-gen object paths=./pkg/apis/...
	@$(GO_TOOL) defaulter-gen --output-file zz_generated.defaults.go ./pkg/apis/...
	@$(GO_TOOL) register-gen --output-file zz_generated.register.go ./pkg/apis/...

.PHONY: generate-operator-extension
generate-operator-extension:
	@$(GO_TOOL) extension-generator \
		--name $(EXTENSION_NAME) \
		--component-category extension \
		--provider-type example \
		--destination $(SRC_ROOT)/examples/operator-extension/base/extension.yaml \
		--extension-oci-repository $(IMAGE):$(IMAGE_TAG)
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/operator-extension

.PHONY: check-helm
check-helm:
	@$(GO_TOOL) helm lint $(SRC_ROOT)/charts
	@$(GO_TOOL) helm template $(SRC_ROOT)/charts | \
		$(GO_TOOL) kubeconform \
			$(KUBECONFORM_OPTS) \
			-schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'

.PHONY: check-examples
check-examples:
	@echo "Checking example resources ..."
	@$(GO_TOOL) kubeconform \
		$(KUBECONFORM_OPTS) \
		-schema-location "$(SRC_ROOT)/test/schemas/{{.Group}}/{{.ResourceAPIVersion}}/{{.ResourceKind}}.json" \
		./examples
	@echo "Checking operator extension resource ..."
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/operator-extension | \
		$(GO_TOOL) kubeconform \
			$(KUBECONFORM_OPTS) \
			-schema-location "$(SRC_ROOT)/test/schemas/{{.Group}}/{{.ResourceAPIVersion}}/{{.ResourceKind}}.json"

.PHONY: kind-load-image
kind-load-image:
	@$(MAKE) docker-build
	@kind load docker-image --name $(KIND_CLUSTER) $(IMAGE):$(IMAGE_TAG)

.PHONY: helm-load-chart
helm-load-chart:
	@$(GO_TOOL) helm package $(SRC_ROOT)/charts --version $(EXTENSION_VERSION)
	@$(GO_TOOL) helm push --plain-http $(EXTENSION_NAME)-$(EXTENSION_VERSION).tgz oci://$(LOCAL_REGISTRY)/helm-charts
	@rm -f $(EXTENSION_NAME)-$(EXTENSION_VERSION).tgz

.PHONY: update-version-tags
update-version-tags:
	@env version=$(EXTENSION_VERSION) \
		$(GO_TOOL) yq -i '.version = env(version)' $(SRC_ROOT)/charts/Chart.yaml
	@env image=$(IMAGE) tag=$(IMAGE_TAG) \
		$(GO_TOOL) yq -i '(.image.repository = env(image)) | (.image.tag = env(tag))' $(SRC_ROOT)/charts/values.yaml
	@env oci_charts=$(LOCAL_REGISTRY)/helm-charts/$(EXTENSION_NAME):$(EXTENSION_VERSION) \
		$(GO_TOOL) yq -i '.helm.ociRepository.ref = env(oci_charts)' $(SRC_ROOT)/examples/dev-setup/controllerdeployment.yaml
	@env oci_charts=$(LOCAL_REGISTRY)/helm-charts/$(EXTENSION_NAME):$(EXTENSION_VERSION) \
		$(GO_TOOL) yq -i '.spec.deployment.extension.helm.ociRepository.ref = env(oci_charts)' $(SRC_ROOT)/examples/operator-extension/base/extension.yaml

.PHONY: deploy
deploy: generate update-version-tags kind-load-image helm-load-chart
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/dev-setup | \
		kubectl apply -f -

.PHONY: undeploy
undeploy:
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/dev-setup | \
		kubectl delete --ignore-not-found=true -f -
