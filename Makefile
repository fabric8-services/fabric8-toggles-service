SOURCE_DIR ?= .
SOURCES := $(shell find $(SOURCE_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)
VENDOR_DIR=vendor
LDFLAGS := -w
BINARY := fabric8-toggles-service

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

.PHONY: clean
## clean build dependencies.
clean:
	rm -rf $(VENDOR_DIR)
	rm -rf glide.lock

.PHONY: deps
## Download build dependencies.
deps: $(VENDOR_DIR)

.PHONY: sysdeps
## Install Glide.
sysdeps:
	go get -u github.com/Masterminds/glide

.PHONY: build
## Build fabric8-toggles-service.
build: deps format-go-code
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY)

.PHONY: run
## Run fabric8-toggles-service.
run: build
	bin/$(BINARY) --config config.yaml
