PROJECT_NAME = fabric8-toggles-service
REGISTRY_URI = push.registry.devshift.net
REGISTRY_NS = fabric8-services
REGISTRY_IMAGE = ${PROJECT_NAME}
REGISTRY_URL = ${REGISTRY_URI}/${REGISTRY_NS}/${REGISTRY_IMAGE}
PACKAGE_NAME := github.com/fabric8-services/${PROJECT_NAME}
SOURCE_DIR ?= .
SOURCES := $(shell find $(SOURCE_DIR) -type d \( -name vendor -o -name .glide \) -prune -o -name '*.go' -print)
VENDOR_DIR=vendor
LDFLAGS := -w
BINARY := ${PROJECT_NAME}
DESIGN_DIR=design
DESIGNS := $(shell find $(SOURCE_DIR)/$(DESIGN_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)
GOAGEN_BIN=$(VENDOR_DIR)/github.com/goadesign/goa/goagen/goagen
CUR_DIR=$(shell pwd)
INSTALL_PREFIX=$(CUR_DIR)/bin
BUILD_DIR = out

# This pattern excludes some folders from the coverage calculation (see grep -v)
ALL_PKGS_EXCLUDE_PATTERN = 'vendor\|app\|tool\/cli\|design\|client\|test'


IMAGE_TAG ?= $(shell git rev-parse HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
IMAGE_TAG := $(IMAGE_TAG)-dirty
endif
BUILD_TIME=`date -u '+%Y-%m-%dT%H:%M:%SZ'`
# Pass in build time variables to main
LDFLAGS=-ldflags "-X ${PACKAGE_NAME}/controller.Commit=${IMAGE_TAG} -X ${PACKAGE_NAME}/controller.BuildTime=${BUILD_TIME}"

.DEFAULT_GOAL := help

GOANALYSIS_DIRS=$(shell go list -f {{.Dir}} ./... | grep -vEf .goanalysisignore)
GOANALYSIS_PKGS=$(shell go list -f {{.ImportPath}} ./... | grep -vEf .goanalysisignore)
GOANALYSIS_FILES=$(shell find  . -name '*.go' | grep -vEf .goanalysisignore)
# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

all: tools generate fmtcheck test image ## Compiles binary and runs format and style checks

$(BUILD_DIR): 
	mkdir $(BUILD_DIR)

$(BUILD_DIR)/$(REGISTRY_IMAGE): vendor $(BUILD_DIR) # Builds the Linux binary for the container image into $BUILD_DIR
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(REGISTRY_IMAGE)

image: $(BUILD_DIR)/$(REGISTRY_IMAGE) ## Builds the container image using the binary compiled for Linux
	docker build -t $(REGISTRY_URL) -f Dockerfile .

push: image ## Pushes the container image to the registry
	$(call check_defined, REGISTRY_USER, "You need to pass the registry user via REGISTRY_USER.")
	$(call check_defined, REGISTRY_PASSWORD, "You need to pass the registry password via REGISTRY_PASSWORD.")
	docker login -u $(REGISTRY_USER) -p $(REGISTRY_PASSWORD) $(REGISTRY_URI)
	docker push $(REGISTRY_URL):latest
	docker tag $(REGISTRY_URL):latest $(REGISTRY_URL):$(IMAGE_TAG)
	docker push $(REGISTRY_URL):$(IMAGE_TAG)

tools: tools.timestamp

tools.timestamp:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	@touch tools.timestamp

vendor: tools.timestamp ## Runs dep to vendor project dependencies
	$(GOPATH)/bin/dep ensure -v

.PHONY: test
test: vendor ## Runs unit tests
	$(eval TEST_PACKAGES:=$(shell go list ./... | grep -v $(ALL_PKGS_EXCLUDE_PATTERN)))
	go test -v $(TEST_PACKAGES)


.PHONY: fmtcheck
fmtcheck: ## Runs gofmt and returns error in case of violations
	@rm -f /tmp/gofmt-errors
	@gofmt -s -l ${GOANALYSIS_FILES} 2>&1 \
		| tee /tmp/gofmt-errors \
		| read \
	&& echo "ERROR: These files differ from gofmt's style (run 'make fmt' to fix this):" \
	&& cat /tmp/gofmt-errors \
	&& exit 1 \
	|| true

.PHONY: fmt
fmt: ## Runs gofmt and formats code violations
	@gofmt -s -w -l ${GOANALYSIS_FILES} 2>&1

.PHONY: vet
vet: ## Runs 'go vet' for common coding mistakes
	@go vet $(GOANALYSIS_PKGS)

.PHONY: lint
lint: ## Runs golint
	@out="$$(golint $(GOANALYSIS_PKGS))"; \
	if [ -n "$$out" ]; then \
		echo "$$out"; \
		exit 1; \
	fi

$(GOAGEN_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v

.PHONY: generate
generate: $(DESIGNS) $(GOAGEN_BIN) $(VENDOR_DIR) ## Generate GOA sources. Only necessary after clean of if changed `design` folder.
	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg app
	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-auth/design --notool --pkg authservice -o auth

.PHONY: run
run: build ## Run fabric8-toggles-service.
	bin/$(BINARY) --config config.yaml

# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

# Keep this "clean" target here at the bottom
.PHONY: clean
clean: $(CLEAN_TARGETS) ## Runs all clean-* targets.

CLEAN_TARGETS += clean-vendor
.PHONY: clean-vendor
clean-vendor: ## clean build dependencies.
	rm -rf $(VENDOR_DIR)
	rm -f tools.timestamp

CLEAN_TARGETS += clean-generated
.PHONY: clean-generated 
clean-generated: ## Removes all generated code.
	-rm -rf ./app
	-rm -rf ./client/
	-rm -rf ./swagger/
	-rm -rf ./tool/cli/
	-rm -rf ./auth

CLEAN_TARGETS += clean-artifacts
.PHONY: clean-artifacts 
clean-artifacts: ## Removes the ./bin directory.
	-rm -rf $(INSTALL_PREFIX)

CLEAN_TARGETS += clean-object-files
.PHONY: clean-object-files 
clean-object-files: ## Runs go clean to remove any executables or other object files.
	go clean ./...

.PHONY: help
help: ## Prints this help
	@grep -E '^[^.]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'








