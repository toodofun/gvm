NAME ?= gvm
OUT_DIR ?= bin
OS_LIST = linux darwin windows
ARCH_LIST = amd64 arm64

VERSION ?= $(shell cat VERSION)

GO_FLAGS ?= -s -w -X 'gvm/core.Version=$(VERSION)'

.PHONY: help
help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: deps
deps:
	@go mod download

.PHONY: release
release:
	@docker buildx build -f Dockerfile --output type=local,dest=bin .

.PHONY: test
test: deps  ## Run unit tests
	go test $(shell go list ./... | grep -v /docs) -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: all
all: build  ## build all

# Create output directory
$(OUT_DIR): deps
	@mkdir -p $(OUT_DIR)

# Build targets for each OS/ARCH combination
.PHONY: build
build: $(OUT_DIR)
	@echo "Building $(NAME) with GO_FLAGS=$(GO_FLAGS)"
	@$(foreach os,$(OS_LIST),\
		$(foreach arch,$(ARCH_LIST),\
			$(call build_target,$(os),$(arch))\
		)\
	)

# Function to build for specific OS/ARCH
define build_target
	@echo "Building $(1)-$(2)"
	$(if $(filter windows,$(1)),\
		GOOS=$(1) GOARCH=$(2) go build -ldflags "$(GO_FLAGS)" -o $(OUT_DIR)/$(NAME)-$(1)-$(2).exe . && \
		cd $(OUT_DIR) && \
		tar zcvf $(NAME)-$(1)-$(2).tar.gz $(NAME)-$(1)-$(2).exe && \
		rm $(NAME)-$(1)-$(2).exe && \
		cd - > /dev/null,\
		GOOS=$(1) GOARCH=$(2) go build -ldflags "$(GO_FLAGS)" -o $(OUT_DIR)/$(NAME)-$(1)-$(2) . && \
		cd $(OUT_DIR) && \
		tar zcvf $(NAME)-$(1)-$(2).tar.gz $(NAME)-$(1)-$(2) && \
		rm $(NAME)-$(1)-$(2) && \
		cd - > /dev/null\
	)
endef

# Build for specific OS
.PHONY: build-linux build-darwin build-windows
build-linux: $(OUT_DIR)
	@echo "Building $(NAME) for Linux"
	@$(foreach arch,$(ARCH_LIST),$(call build_target,linux,$(arch)))

build-darwin: $(OUT_DIR)
	@echo "Building $(NAME) for macOS"
	@$(foreach arch,$(ARCH_LIST),$(call build_target,darwin,$(arch)))

build-windows: $(OUT_DIR)
	@echo "Building $(NAME) for Windows"
	@$(foreach arch,$(ARCH_LIST),$(call build_target,windows,$(arch)))

# Build for specific architecture
.PHONY: build-amd64 build-arm64
build-amd64: $(OUT_DIR)
	@echo "Building $(NAME) for amd64"
	@$(foreach os,$(OS_LIST),$(call build_target,$(os),amd64))

build-arm64: $(OUT_DIR)
	@echo "Building $(NAME) for arm64"
	@$(foreach os,$(OS_LIST),$(call build_target,$(os),arm64))

# Clean build directory
.PHONY: clean
clean: ## clean
	@echo "Cleaning $(OUT_DIR)"
	@rm -rf $(OUT_DIR)
	@echo "Cleaning test data"
	@rm -rf coverage.out
