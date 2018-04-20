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
MINIMOCK_BIN=$(VENDOR_DIR)/github.com/gojuno/minimock/cmd/minimock/minimock
CUR_DIR=$(shell pwd)
BUILD_DIR = bin
F8_AUTH_URL ?= https://auth.prod-preview.openshift.io
F8_TOGGLES_URL ?= "http://toggles:4242/api"
FABRIC8_MARKER=.fabric8
FABRIC8_PROJECT=fabric8

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

all: tools.timestamp generate fmtcheck test-coverage image ## Compiles binary and runs format and style checks

$(BUILD_DIR): 
	mkdir $(BUILD_DIR)

.PHONY: build
build: deps generate $(BUILD_DIR) # Builds the Linux binary for the container image into $BUILD_DIR
	go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)

.PHONY: build-linux
build-linux: deps generate $(BUILD_DIR) # Builds the Linux binary for the container image into $BUILD_DIR
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)

image: clean-artifacts build-linux
	docker build -t $(REGISTRY_URL) \
	  --build-arg BINARY=$(BUILD_DIR)/$(PROJECT_NAME) \
	  -f Dockerfile .

image-minishift: clean-artifacts build-linux
	@eval $$(minishift docker-env) && docker build -t $(REGISTRY_URL) \
	  --build-arg BINARY=$(BUILD_DIR)/$(PROJECT_NAME) \
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
	@echo "checking dependencies..."
	$(GOPATH)/bin/dep ensure -v 


.PHONY: test
test: deps ## Runs unit tests without code coverage
	$(eval TEST_PACKAGES:=$(shell go list ./... | grep -v $(ALL_PKGS_EXCLUDE_PATTERN)))
	go test -v $(TEST_PACKAGES)

.PHONY: test-coverage
test-coverage: deps tmp ## Runs unit tests with coverage support
	@-rm -f coverage.txt
	@echo "running tests with code coverage"
	@for d in `go list ./... | grep -v $(ALL_PKGS_EXCLUDE_PATTERN)`; do \
		echo "running tests with code coverage in pkg $$d" ; \
		go test -coverprofile=./tmp/profile.out -covermode=atomic $$d ; \
		if [ -e ./tmp/profile.out ]; \
			then \
			cat ./tmp/profile.out >> coverage.txt && rm ./tmp/profile.out ; \
		fi \
	done

tmp:
	@mkdir tmp

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

.PHONY: generate
generate: generate-goa generate-minimock ## Generate GOA and Minimock sources.

$(GOAGEN_BIN):
	@echo "building the goagen binary..."
	@cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v

.PHONY: generate-goa
generate-goa: deps $(DESIGNS) $(GOAGEN_BIN) deps ## Generate GOA sources. Only necessary after clean or if changes occurred in `design` folder.
	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg app
	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-auth/design --notool --pkg client -o auth
	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} --pkg-path=${PACKAGE_NAME}/goasupport/conditional_request --out app

$(MINIMOCK_BIN):
	@echo "building the minimock binary..."
	@cd $(VENDOR_DIR)/github.com/gojuno/minimock/cmd/minimock && go build -v minimock.go

.PHONY: generate-minimock
generate-minimock: deps $(MINIMOCK_BIN) ## Generate Minimock sources. Only necessary after clean or if changes occurred in interfaces.
	@echo "Generating mocks..."
	@-mkdir -p test/token
	@$(MINIMOCK_BIN) -i github.com/fabric8-services/fabric8-toggles-service/vendor/github.com/fabric8-services/fabric8-auth/token.Parser -o ./test/token/parser_mock.go -t ParserMock
	@-mkdir -p test/featuretoggles
	@$(MINIMOCK_BIN) -i github.com/fabric8-services/fabric8-toggles-service/featuretoggles.UnleashClient -o ./test/featuretoggles/unleashclient_mock.go -t UnleashClientMock
	@$(MINIMOCK_BIN) -i github.com/fabric8-services/fabric8-toggles-service/featuretoggles.Client -o ./test/featuretoggles/toggles_client_wrapper_mock.go -t ClientMock

.PHONY: run
run: build ## Run fabric8-toggles-service.
	$(BUILD_DIR)/$(PROJECT_NAME) --config config.yaml

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

## the '-' at the beginning of the line will ignore failure of `oc project` if the project already exists.
$(FABRIC8_MARKER):
	@-oc new-project ${FABRIC8_PROJECT} 2>/dev/null
	@touch $@

.PHONY: push-minishift
push-minishift: minishift-login minishift-registry-login image-minishift $(FABRIC8_MARKER)
	eval $$(minishift docker-env) && docker login -u developer -p $(shell oc whoami -t) $(shell minishift openshift registry) && docker tag ${REGISTRY_URI}/${REGISTRY_NS}/${REGISTRY_IMAGE}  $(shell minishift openshift registry)/${FABRIC8_PROJECT}/${REGISTRY_IMAGE}:latest
	eval $$(minishift docker-env) && docker login -u developer -p $(shell oc whoami -t) $(shell minishift openshift registry) && docker push $(shell minishift openshift registry)/${FABRIC8_PROJECT}/${REGISTRY_IMAGE}:latest

.PHONY: deploy-minishift
deploy-minishift: push-minishift ## deploy toggles server on minishift
	curl https://raw.githubusercontent.com/xcoulon/fabric8-minishift/master/toggles-db.yml -o ./minishift/toggles-db.yml
	kedge apply -f ./minishift/toggles-db.yml
	curl https://raw.githubusercontent.com/xcoulon/fabric8-minishift/master/toggles.yml -o ./minishift/toggles.yml
	kedge apply -f ./minishift/toggles.yml
	curl https://raw.githubusercontent.com/xcoulon/fabric8-minishift/master/toggles-service.yml -o ./minishift/toggles-service.yml
	F8_AUTH_URL=$(F8_AUTH_URL) F8_TOGGLES_URL=$(F8_TOGGLES_URL) kedge apply -f ./minishift/toggles-service.yml

.PHONY: clean-minishift
clean-minishift: minishift-login ## removes the fabric8 project on Minishift
	rm -rf $(FABRIC8_MARKER)
	oc project fabric8 && oc delete project fabric8


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
	-rm -rf ./swagger/
	-rm -rf ./auth/client

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








