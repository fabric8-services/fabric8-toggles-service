PACKAGE_NAME := github.com/fabric8-services/fabric8-toggles-service
SOURCE_DIR ?= .
SOURCES := $(shell find $(SOURCE_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)
VENDOR_DIR=vendor
LDFLAGS := -w
BINARY := fabric8-toggles-service
DESIGN_DIR=design
DESIGNS := $(shell find $(SOURCE_DIR)/$(DESIGN_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)
GOAGEN_BIN=$(VENDOR_DIR)/github.com/goadesign/goa/goagen/goagen
CUR_DIR=$(shell pwd)
INSTALL_PREFIX=$(CUR_DIR)/bin

# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

$(GOAGEN_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v

# If nothing was specified, run all targets as if in a fresh clone
.PHONY: all
## Default target - fetch dependencies, generate code and build.
all: sysdeps deps generate build

.PHONY: help
# Based on https://gist.github.com/rcmachado/af3db315e31383502660
## Display this help text.
help:/
	$(info Available targets)
	$(info -----------------)
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		helpCommand = substr($$1, 0, index($$1, ":")-1); \
		if (helpMessage) { \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub(/##/, "\n                                     ", helpMessage); \
		} else { \
			helpMessage = "(No documentation)"; \
		} \
		printf "%-35s - %s\n", helpCommand, helpMessage; \
		lastLine = "" \
	} \
	{ hasComment = match(lastLine, /^## (.*)/); \
          if(hasComment) { \
            lastLine=lastLine$$0; \
	  } \
          else { \
	    lastLine = $$0 \
          } \
        }' $(MAKEFILE_LIST)

.PHONY: format-go-code
## Formats any go file that differs from gofmt's style
format-go-code:
	gofmt -l -s -w ${SOURCES}

.PHONY: dev
## Start fabric8-toggles-service in development mode.
dev: deps
	echo 'TODO Docker'

$(VENDOR_DIR): glide.yaml
	$(GOPATH)/bin/glide install
	touch $(VENDOR_DIR)

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

.PHONY: deps
## Download build dependencies.
deps: $(VENDOR_DIR)

.PHONY: sysdeps
## Install Glide.
sysdeps:
	go get -u github.com/Masterminds/glide


.PHONY: generate
## Generate GOA sources. Only necessary after clean of if changed `design` folder.
generate: $(DESIGNS) $(GOAGEN_BIN) $(VENDOR_DIR)
	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg app
	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-auth/design --notool --pkg authservice -o auth

.PHONY: regenerate
## Runs the "clean-generated" and the "generate" target
regenerate: clean-generated generate

.PHONY: build
## Build fabric8-toggles-service.
build: deps format-go-code generate
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY)

.PHONY: run
## Run fabric8-toggles-service.
run: build
	bin/$(BINARY) --config config.yaml

# Keep this "clean" target here at the bottom
.PHONY: clean
## Runs all clean-* targets.
clean: $(CLEAN_TARGETS)
