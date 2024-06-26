# SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors.
#
# SPDX-License-Identifier: Apache-2.0

REPO_ROOT         := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
EFFECTIVE_VERSION := $(shell $(REPO_ROOT)/hack/get-version.sh)

CODE_DIRS := $(REPO_ROOT)/cmd/... $(REPO_ROOT)/pkg/... $(REPO_ROOT)/landscaper-cli/... $(REPO_ROOT)/integration-test/...


##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Development

.PHONY: revendor
revendor: ## Runs 'go mod tidy' for all go modules in this repo.
	@$(REPO_ROOT)/hack/revendor.sh

.PHONY: format
format: goimports ## Runs the formatter.
	@@FORMATTER=$(FORMATTER) $(REPO_ROOT)/hack/format.sh $(CODE_DIRS)

.PHONY: check
check: golangci-lint goimports ## Runs linter, 'go vet', and checks if the formatter has been run.
	@LINTER=$(LINTER) FORMATTER=$(FORMATTER) $(REPO_ROOT)/hack/check.sh --golangci-lint-config="$(REPO_ROOT)/.golangci.yaml" $(CODE_DIRS)

.PHONY: verify
verify: check ## Alias for 'make check'.

.PHONY: generate
generate: ## Generates the command reference documentation.
	@go run "$(REPO_ROOT)/hack/generate-docs/main.go" "$(REPO_ROOT)/docs/reference"

.PHONY: test
test: ## Runs the tests.
	@go test $(CODE_DIRS)


##@ Build

.PHONY: install-cli
install-cli: ## Installs the CLI.
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) $(REPO_ROOT)/hack/install-cli.sh

.PHONY: cross-build
cross-build: ## Builds the binary for linux/amd64, darwin/amd64, and darwin/arm64.
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) $(REPO_ROOT)/hack/cross-build.sh

.PHONY: component
component: component-build component-push ## Builds the components and pushes them into the registry. To overwrite existing versions, set the env var OVERWRITE_COMPONENTS to anything except 'false' or the empty string.

.PHONY: component-build
component-build: ocm ## Build the components.
	OCM=$(OCM) $(REPO_ROOT)/hack/build-component.sh

.PHONY: component-push
component-push: ocm ## Upload the components into the registry. Must be called after 'make component-build'. To overwrite existing versions, set the env var OVERWRITE_COMPONENTS to anything except 'false' or the empty string.
	OCM=$(OCM) $(REPO_ROOT)/hack/push-component.sh


##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(REPO_ROOT)/bin

## Tool Binaries
FORMATTER ?= $(LOCALBIN)/goimports
LINTER ?= $(LOCALBIN)/golangci-lint
OCM ?= $(LOCALBIN)/ocm

## Tool Versions
# renovate: datasource=github-tags depName=golang/tools
FORMATTER_VERSION ?= v0.21.0
# renovate: datasource=github-releases depName=golangci/golangci-lint
LINTER_VERSION ?= 1.59.1
# renovate: datasource=github-releases depName=open-component-model/ocm
OCM_VERSION ?= 0.11.0

.PHONY: localbin
localbin: ## Creates the local bin folder, if it doesn't exist. Not meant to be called manually, used as requirement for the other tool commands.
	@test -d $(LOCALBIN) || mkdir -p $(LOCALBIN)

.PHONY: goimports
goimports: localbin ## Download goimports locally if necessary. If wrong version is installed, it will be overwritten.
	@test -s $(FORMATTER) && test -s $(LOCALBIN)/goimports_version && cat $(LOCALBIN)/goimports_version | grep -q $(FORMATTER_VERSION) || \
	( echo "Installing goimports $(FORMATTER_VERSION) ..."; \
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(FORMATTER_VERSION) && \
	echo $(FORMATTER_VERSION) > $(LOCALBIN)/goimports_version )

.PHONY: golangci-lint
golangci-lint: localbin ## Download golangci-lint locally if necessary. If wrong version is installed, it will be overwritten.
	@test -s $(LINTER) && $(LINTER) --version | grep -q $(LINTER_VERSION) || \
	( echo "Installing golangci-lint $(LINTER_VERSION) ..."; \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) v$(LINTER_VERSION) )

.PHONY: ocm
ocm: localbin ## Install OCM CLI if necessary. If wrong version is installed, it will be overwritten.
	@test -s $(OCM) && $(OCM) --version | grep -q $(subst v,,$(OCM_VERSION)) || \
	( echo "Installing OCM tooling $(OCM_VERSION) ..."; \
	curl -sSfL https://ocm.software/install.sh | OCM_VERSION=$(subst v,,$(OCM_VERSION)) bash -s $(LOCALBIN) )
