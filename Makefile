# Copyright 2025 The Toodofun Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http:www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build all by default, even if it's not first
.DEFAULT_GOAL := all

.PHONY: all
all: tidy add-copyright format lint cover build

# ==============================================================================
# Build options

GO := go
OS = linux darwin windows
ARCH_LIST = amd64 arm64
NAME = gvm
ROOT_PACKAGE=github.com/toodofun/gvm
COVERAGE := 20
GOLANG_CI_LINT_VERSION ?= 2.9.0
SHELL := /bin/bash
DOCKER := docker

# Docker command settings
REGISTRY_PREFIX ?= "ghcr.io"
IMAGE ?= "toodofun/gvm"
VERSION ?= $(shell git describe --tags)
TAG := $(REGISTRY_PREFIX)/$(IMAGE):$(VERSION)
DOCKER_FILE ?= "Dockerfile.dev"
DOCKER_BUILD_ARG_RELEASE ?= $(VERSION)
DOCKER_MULTI_ARCH ?= linux/amd64,linux/arm64

# Linux command settings
FIND := find . ! -path './vendor/*'
XARGS := xargs -r
COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(shell cd "$(COMMON_SELF_DIR)" && pwd -P)
endif

# Create output directory
ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/_output
$(shell mkdir -p "$(OUTPUT_DIR)")
endif

ifeq ($(origin BIN_DIR),undefined)
BIN_DIR := $(ROOT_DIR)/bin
$(shell mkdir -p $(BIN_DIR))
endif

GO_LDFLAGS := $(shell $(GO) run "$(ROOT_DIR)/scripts/gen-ldflags.go")
GO_BUILD_FLAGS = --ldflags "$(GO_LDFLAGS)"

# ==============================================================================
# Includes

include scripts/Makefile.tools.mk

# ==============================================================================
# Targets

## lint: Check syntax and styling of go sources.
.PHONY: lint
lint: tools.verify.local.golangci-lint
	@echo "===========> Run golangci to lint source codes"
	@$(BIN_DIR)/golangci-lint run -c $(ROOT_DIR)/.golangci.yml $(ROOT_DIR)/...

## test: Run unit test.
.PHONY: test
test: tools.verify.go-junit-report
	@echo "===========> Run unit test"
	@set -o pipefail;$(GO) test -tags=test $(shell go list ./...) -race -cover -coverprofile="$(OUTPUT_DIR)/coverage.out" \
		-timeout=10m -shuffle=on -short \
	@$(GO) tool cover -html="$(OUTPUT_DIR)/coverage.out" -o "$(OUTPUT_DIR)/coverage.html"
	@$(GO) tool cover -func="$(OUTPUT_DIR)/coverage.out"

## cover: Run unit test and get test coverage.
.PHONY: cover
cover: test
	@$(GO) tool cover -func="$(OUTPUT_DIR)/coverage.out" | \
		awk -v target=$(COVERAGE) -f "$(ROOT_DIR)/scripts/coverage.awk"

## format: Gofmt (reformat) package sources (exclude vendor dir if existed).
.PHONY: format
format: tools.verify.golines tools.verify.goimports
	@echo "===========> Formating codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -s -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(ROOT_PACKAGE)
	@$(FIND) -type f -name '*.go' | $(XARGS) golines -w --max-len=120 --reformat-tags --shorten-comments --ignore-generated .
	@$(GO) mod edit -fmt

## verify-copyright: Verify the boilerplate headers for all files.
.PHONY: verify-copyright
verify-copyright: tools.verify.licctl
	@echo "===========> Verifying the boilerplate headers for all files"
	@licctl --check -f "$(ROOT_DIR)/scripts/boilerplate.txt" "$(ROOT_DIR)" --skip-dirs=_output,testdata,.github,.idea

## add-copyright: Ensures source code files have copyright license headers.
.PHONY: add-copyright
add-copyright: tools.verify.licctl
	@licctl -v -f "$(ROOT_DIR)/scripts/boilerplate.txt" "$(ROOT_DIR)" --skip-dirs=_output,testdata,.github,.idea

## build: generate releases for unix and windows systems
.PHONY: build
build: clean tidy
	@echo "===========> build flags=$(GO_BUILD_FLAGS)"
	@for arch in $(ARCH_LIST); do \
		for os in $(OS); do \
			echo "Building $$os-$$arch"; \
			ext=""; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/$(NAME)-$$os-$$arch$$ext .; \
		done \
	done

## image.build.%: Docker build for image
.PHONY: image.build.%
image.build.%: build
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@echo "===========> Building and Pushing $(TAG) for $(OS) $(ARCH) $(ROOT_DIR)/$(DOCKER_FILE)"
	@${DOCKER} buildx build --push -t $(TAG) --build-arg TARGETARCH=${OS}-${ARCH} --build-arg RELEASE=${DOCKER_BUILD_ARG_RELEASE} \
 		--platform $(DOCKER_MULTI_ARCH) -f "$(ROOT_DIR)/$(DOCKER_FILE)" "$(ROOT_DIR)"

## image.push: Push docker mirror repository
.PHONY: image.push
image.push:
	@echo "===========> Pushing image $(TAG)"
	@${DOCKER} push $(TAG)

## tools: Install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

## deps: Download all Go module dependencies listed in go.mod
.PHONY: deps
deps:
	@$(GO) mod download

## check-updates: Check for available updates of direct Go module dependencies
.PHONY: check-updates
check-updates: tools.verify.go-mod-outdated
	@$(GO) list -u -m -json all | go-mod-outdated -update -direct

## clean: Install dependent tools.
.PHONY: clean
clean: ## Remove building artifacts
	@echo "===========> Cleaning all build output"
	rm -rf $(OUTPUT_DIR)/*

## tidy: Clean up go.mod and go.sum by removing unused dependencies and adding missing ones
.PHONY: tidy
tidy:
	@$(GO) mod tidy

## help: Show this help info.
.PHONY: help
help: Makefile
	@printf "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'