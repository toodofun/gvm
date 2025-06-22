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
all: build

NAME ?= gvm
OS_LIST = linux darwin windows
ARCH_LIST = amd64 arm64
SHELL := /bin/bash

GO := go
ROOT_DIR=.
ROOT_PACKAGE=github.com/toodofun/gvm

# Linux command settings
FIND := find .
XARGS := xargs -r

VERSION ?= $(shell cat VERSION)

GO_FLAGS ?= -s -w -X 'gvm/core.Version=$(VERSION)'

# Create output directory
ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/_output
$(shell mkdir -p $(OUTPUT_DIR))
endif

# Minimum test coverage
ifeq ($(origin COVERAGE),undefined)
# prod
#COVERAGE := 60
# develop
COVERAGE := 0.0
endif

include scripts/Makefile.tools.mk

.PHONY: deps
deps:
	@go mod download

.PHONY: release
release:
	@docker buildx build -f Dockerfile --output type=local,dest=bin .

## lint: Check syntax and styling of go sources.
.PHONY: lint
lint: tools.verify.golangci-lint
	@echo "===========> Run golangci to lint source codes"
	@golangci-lint run -c $(ROOT_DIR)/.golangci.yml $(ROOT_DIR)/...

## test: Run unit test.
.PHONY: test
test: tools.verify.go-junit-report
	@echo "===========> Run unit test"
	@set -o pipefail;$(GO) test -tags=test $(shell go list ./... | grep -v view) -race -cover -coverprofile=$(OUTPUT_DIR)/coverage.out \
		-timeout=10m -shuffle=on -short \
	@$(GO) tool cover -html=$(OUTPUT_DIR)/coverage.out -o $(OUTPUT_DIR)/coverage.html
	@$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out

## cover: Run unit test and get test coverage.
.PHONY: cover
cover: test
	@$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out | \
		awk -v target=$(COVERAGE) -f $(ROOT_DIR)/scripts/coverage.awk

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
	@licctl --check -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) --skip-dirs=_output,testdata,.github,.idea

## add-copyright: Ensures source code files have copyright license headers.
.PHONY: add-copyright
add-copyright: tools.verify.licctl
	@licctl -v -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) --skip-dirs=_output,testdata,.github,.idea

## tools: install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

## Build targets for each OS/ARCH combination
.PHONY: build
build:
	@echo "Building $(NAME) with GO_FLAGS=$(GO_FLAGS)"
	@$(foreach os,$(OS_LIST),\
		$(foreach arch,$(ARCH_LIST),\
			$(call build_target,$(os),$(arch))\
		)\
	)

## Function to build for specific OS/ARCH
define build_target
	@echo "Building $(1)-$(2)"
	$(if $(filter windows,$(1)),\
		GOOS=$(1) GOARCH=$(2) go build -ldflags "$(GO_FLAGS)" -o $(OUTPUT_DIR)/$(NAME)-$(1)-$(2).exe . && \
		cd $(OUTPUT_DIR) && \
		tar zcvf $(NAME)-$(1)-$(2).tar.gz $(NAME)-$(1)-$(2).exe && \
		rm $(NAME)-$(1)-$(2).exe && \
		cd - > /dev/null,\
		GOOS=$(1) GOARCH=$(2) go build -ldflags "$(GO_FLAGS)" -o $(OUTPUT_DIR)/$(NAME)-$(1)-$(2) . && \
		cd $(OUTPUT_DIR) && \
		tar zcvf $(NAME)-$(1)-$(2).tar.gz $(NAME)-$(1)-$(2) && \
		rm $(NAME)-$(1)-$(2) && \
		cd - > /dev/null\
	)
endef

## Build for specific OS
.PHONY: build-linux build-darwin build-windows
build-linux:
	@echo "Building $(NAME) for Linux"
	@$(foreach arch,$(ARCH_LIST),$(call build_target,linux,$(arch)))

build-darwin:
	@echo "Building $(NAME) for macOS"
	@$(foreach arch,$(ARCH_LIST),$(call build_target,darwin,$(arch)))

build-windows:
	@echo "Building $(NAME) for Windows"
	@$(foreach arch,$(ARCH_LIST),$(call build_target,windows,$(arch)))

## Build for specific architecture
.PHONY: build-amd64 build-arm64
build-amd64:
	@echo "Building $(NAME) for amd64"
	@$(foreach os,$(OS_LIST),$(call build_target,$(os),amd64))

build-arm64:
	@echo "Building $(NAME) for arm64"
	@$(foreach os,$(OS_LIST),$(call build_target,$(os),arm64))

## Clean build directory
.PHONY: clean
clean:
	@echo "Cleaning $(OUTPUT_DIR)"
	@rm -rf $(OUTPUT_DIR)
	@echo "Cleaning test data"
	@rm -rf coverage.out

.PHONY: tidy
tidy:
	@$(GO) mod tidy

.PHONY: help
help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'