.PHONY: test build

GOBIN = $(dir $(realpath $(firstword $(MAKEFILE_LIST))))build/bin
PROJECT_NAME=keycard-cli
GO_PROJECT_PATH=github.com/status-im/$(PROJECT_NAME)
BIN_NAME=keycard
DOCKER_IMAGE_NAME=keycard

build:
	go build -i -o $(GOBIN)/$(BIN_NAME) -v .
	@echo "Compilation done."
	@echo "Run \"./build/bin/$(BIN_NAME) -h\" to view available commands."

test:
	go test -v ./...

build-docker-image:
	docker build -t $(DOCKER_IMAGE_NAME) -f _assets/Dockerfile .

build-platforms:
	xgo -image $(DOCKER_IMAGE_NAME) --dest $(GOBIN) --targets=linux/amd64,windows/amd64 .

