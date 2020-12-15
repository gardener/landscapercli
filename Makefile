# SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors.
#
# SPDX-License-Identifier: Apache-2.0

REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION                                        := $(shell cat $(REPO_ROOT)/VERSION)
EFFECTIVE_VERSION                              := $(VERSION)-$(shell git rev-parse HEAD)


.PHONY: install-requirements
install-requirements:
	@go install $(REPO_ROOT)/vendor/github.com/ahmetb/gen-crd-api-reference-docs
	@go install $(REPO_ROOT)/vendor/github.com/golang/mock/mockgen
	@$(REPO_ROOT)/hack/install-requirements.sh

.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(REPO_ROOT)/pkg $(REPO_ROOT)/cmd

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(REPO_ROOT)/cmd/... $(REPO_ROOT)/pkg/...

.PHONY: test
test:
	@go test $(REPO_ROOT)/cmd/... $(REPO_ROOT)/pkg/...

.PHONY: verify
verify: check

.PHONY: install-cli
install-cli:
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) ./hack/install-cli.sh
