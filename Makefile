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

all: tools build test fmtcheck vet image ## Compiles binary and runs format and style checks

build: vendor generate ## Builds the binary into $GOPATH/bin
	go install $(LDFLAGS) ./cmd/fabric8-jenkins-idler

$(BUILD_DIR): 
	mkdir $(BUILD_DIR)

$(BUILD_DIR)/$(REGISTRY_IMAGE): vendor $(BUILD_DIR) # Builds the Linux binary for the container image into $BUILD_DIR
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build  -ldflags="$(LD_FLAGS)" -o $(BUILD_DIR)/$(REGISTRY_IMAGE) ./cmd/fabric8-jenkins-idler

image: $(BUILD_DIR)/$(REGISTRY_IMAGE) ## Builds the container image using the binary compiled for Linux
	docker build -t $(REGISTRY_URL) -f Dockerfile.deploy .

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
	go test $(TEST_PACKAGES)


.PHONY: fmtcheck
fmtcheck: ## Runs gofmt and returns error in case of violations
	@gofmt -l -s $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: fmt
fmt: ## Runs gofmt and formats code violations
	@gofmt -l -s -w $(SOURCE_DIRS)

.PHONY: vet
vet: ## Runs 'go vet' for common coding mistakes
	@go vet $(PACKAGES)

.PHONY: lint
lint: ## Runs golint
	@out="$$(golint $(PACKAGES))"; \
	if [ -n "$$out" ]; then \
		echo "$$out"; \
		exit 1; \
	fi

$(GOAGEN_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v

.PHONY: generate
## Generate GOA sources. Only necessary after clean of if changed `design` folder.
generate: $(DESIGNS) $(GOAGEN_BIN) $(VENDOR_DIR)
	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg app
	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-auth/design --notool --pkg authservice -o auth

.PHONY: run
## Run fabric8-toggles-service.
run: build
	bin/$(BINARY) --config config.yaml


# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

.PHONY: help
help: ## Prints this help
	@grep -E '^[^.]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'


# Keep this "clean" target here at the bottom
.PHONY: clean
## Runs all clean-* targets.
clean: $(CLEAN_TARGETS)

CLEAN_TARGETS += clean-deps
.PHONY: clean-deps
## clean build dependencies.
clean-deps:
	rm -rf $(VENDOR_DIR)

CLEAN_TARGETS += clean-generated
.PHONY: clean-generated
## Removes all generated code.
clean-generated:
	-rm -rf ./app
	-rm -rf ./client/
	-rm -rf ./swagger/
	-rm -rf ./tool/cli/
	-rm -rf ./feature/feature
	-rm -rf ./auth

CLEAN_TARGETS += clean-artifacts
.PHONY: clean-artifacts
## Removes the ./bin directory.
clean-artifacts:
	-rm -rf $(INSTALL_PREFIX)

CLEAN_TARGETS += clean-object-files
.PHONY: clean-object-files
## Runs go clean to remove any executables or other object files.
clean-object-files:
	go clean ./...











