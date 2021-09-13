export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)
export INSTALL_FLAG=

# Shell to use for running scripts
export SHELL := /bin/bash
# Get docker path or an empty string
DOCKER := $(shell command -v docker)

DOCKER_IMAGE = objectrocket/sensu-operator
# allow builds without tags
IMAGE_VERSION ?= latest
VERSION ?= $(shell git describe --tags || git symbolic-ref -q --short HEAD)

# Test if the dependencies we need to run this Makefile are installed
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif

.PHONY: all
all: build container

.PHONY: build
build:
	@go build -ldflags="-X github.com/objectrocket/sensu-operator/version.Version=$VERSION" -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o _output/sensu-operator cmd/operator/main.go

.PHONY: test
test:
	@hack/test

.PHONY: unittest
unittest:
	@hack/unit_test

.PHONY: clean
clean:
	@go clean

.PHONY: dep
dep:
	@go mod download
	@go mod tidy

docker-build: deps-development
	docker build --build-arg APPVERSION=$(IMAGE_VERSION) -t $(DOCKER_IMAGE):$(IMAGE_VERSION) .

docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(IMAGE_VERSION)
