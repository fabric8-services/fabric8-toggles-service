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
DESIGN_DIR=design
DESIGNS := $(shell find $(SOURCE_DIR)/$(DESIGN_DIR) -path $(SOURCE_DIR)/$(VENDOR_DIR) -prune -o -name '*.go' -print)
GOAGEN_BIN=$(VENDOR_DIR)/github.com/goadesign/goa/goagen/goagen
CUR_DIR=$(shell pwd)
BUILD_DIR = bin
F8_AUTH_URL ?= https://auth.prod-preview.openshift.io
F8_TOGGLES_URL ?= "http://toggles:4242/api"
F8_KEYCLOAK_URL ?= "https://sso.prod-preview.openshift.io"
fabric8=fabric8

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

all: tools.timestamp generate fmtcheck test image ## Compiles binary and runs format and style checks

$(BUILD_DIR): 
	mkdir $(BUILD_DIR)

.PHONY: build
build: deps generate $(BUILD_DIR) # Builds the Linux binary for the container image into $BUILD_DIR
	go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(REGISTRY_IMAGE)

.PHONY: build-linux
build-linux: deps generate $(BUILD_DIR) # Builds the Linux binary for the container image into $BUILD_DIR
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(REGISTRY_IMAGE)

image: clean-artifacts build-linux ## Builds the container image using the binary compiled for Linux
	docker build -t $(REGISTRY_URL) \
	  --build-arg BINARY=$(BUILD_DIR)/$(REGISTRY_IMAGE) \
	  -f Dockerfile .

push-openshift: image ## Pushes the container image to the OpenShift online registry
	$(call check_defined, REGISTRY_USER, "You need to pass the registry user via REGISTRY_USER.")
	$(call check_defined, REGISTRY_PASSWORD, "You need to pass the registry password via REGISTRY_PASSWORD.")
	docker login -u $(REGISTRY_USER) -p $(REGISTRY_PASSWORD) $(REGISTRY_URI)
	docker push $(REGISTRY_URL):latest
	docker tag $(REGISTRY_URL):latest $(REGISTRY_URL):$(IMAGE_TAG)
	docker push $(REGISTRY_URL):$(IMAGE_TAG)

tools.timestamp:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	@touch tools.timestamp

deps: tools.timestamp $(VENDOR_DIR) ## Runs dep to vendor project dependencies

$(VENDOR_DIR):
	$(GOPATH)/bin/dep ensure -v 


.PHONY: test
test: deps ## Runs unit tests
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

$(GOAGEN_BIN): deps
	cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v

.PHONY: generate
generate: $(DESIGNS) $(GOAGEN_BIN) deps ## Generate GOA sources. Only necessary after clean of if changed `design` folder.
	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg app
	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-auth/design --notool --pkg authservice -o auth

.PHONY: run
run: build ## Run fabric8-toggles-service.
	$(BUILD_DIR)/$(REGISTRY_IMAGE) --config config.yaml

.PHONY: minishift-login
## login to oc minishift
minishift-login:
	@echo "Login to minishift..."
	@oc login --insecure-skip-tls-verify=true -u developer -p developer

.PHONY: minishift-registry-login
## login to the registry in Minishift (to push images)
minishift-registry-login:
	@echo "Login to minishift registry..."
	@eval $$(minishift docker-env) && docker login -u developer -p $(shell oc whoami -t) $(shell minishift openshift registry)

$(fabric8):
	oc new-project fabric8
	touch $@

.PHONY: push-minishift
push-minishift: minishift-login minishift-registry-login image $(fabric8)
	docker tag ${REGISTRY_URI}/${REGISTRY_NS}/${REGISTRY_IMAGE}  $(shell minishift openshift registry)/${fabric8}/${REGISTRY_IMAGE}:latest
	docker push $(shell minishift openshift registry)/${fabric8}/${REGISTRY_IMAGE}:latest

.PHONY: deploy-minishift
deploy-minishift: push-minishift ## deploy toggles server on minishift
	kedge apply -f ./minishift/toggles-db.yml
	kedge apply -f ./minishift/toggles.yml
	F8_AUTH_URL=$(F8_AUTH_URL) F8_KEYCLOAK_URL=$(F8_KEYCLOAK_URL) F8_TOGGLES_URL=$(F8_TOGGLES_URL) kedge apply -f ./minishift/toggles-service.yml

.PHONY: clean-minishift
clean-minishift: minishift-login ## removes the fabric8 project on Minishift
	oc project fabric8 && oc delete project fabric8 && rm -rf $(fabric8)


# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

CLEAN_TARGETS += clean-deps
.PHONY: clean-deps
clean-deps: ## clean vendor dependencies.
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
clean-artifacts: ## Removes the ./out directory.
	-rm -rf $(BUILD_DIR)

CLEAN_TARGETS += clean-object-files
.PHONY: clean-object-files 
clean-object-files: ## Runs go clean to remove any executables or other object files.
	go clean ./...

# Keep this "clean" target here at the bottom (and don't scream)
.PHONY: clean
clean: $(CLEAN_TARGETS) ## Runs all clean-* targets.

.PHONY: help
help: ## Prints this help
	@grep -E '^[^.]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'








