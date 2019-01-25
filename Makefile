export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)
export INSTALL_FLAG=

# Shell to use for running scripts
export SHELL := /bin/bash
# Get docker path or an empty string
DOCKER := $(shell command -v docker)

IMAGE ?= objectrocket/sensu-operator:v0.0.1
DOCKER_IMAGE = objectrocket/sensu-operator

# Test if the dependencies we need to run this Makefile are installed
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif
ifndef IMAGE_VERSION
	@echo "Variable IMAGE_VERSION is required"
	@exit 1
endif
# allow circle to run builds even without a tag
ifeq ($(IMAGE_VERSION),)
	IMAGE_VERSION := latest
endif

.PHONY: all
all: build container

.PHONY: build
build:
	@hack/build/operator/build

.PHONY: test
test:
	@hack/test

.PHONY: unittest
unittest:
	@hack/unit_test

.PHONY: clean
clean:
	@go clean

docker-build: deps-development
	docker build -t $(DOCKER_IMAGE):$(IMAGE_VERSION) .

docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(IMAGE_VERSION)
