.PHONY: test build

# This can be changed by exporting an env variable
XGO_TARGETS ?= linux/amd64,windows/amd64,darwin/amd64

GOBIN = $(dir $(realpath $(firstword $(MAKEFILE_LIST))))build/bin
PROJECT_NAME=keycard-cli
BIN_NAME=keycard

VERSION = $(shell cat VERSION)

export GITHUB_USER ?= status-im
export GITHUB_REPO ?= $(PROJECT_NAME)

export IMAGE_TAG   ?= latest
export IMAGE_NAME  ?= statusteam/keycard-cli-ci:$(IMAGE_TAG)

export GO_PROJECT_PATH ?= github.com/$(GITHUB_USER)/$(GITHUB_REPO)

deps:
	go get github.com/karalabe/xgo
	go get github.com/aktau/github-release

build:
	go build -i -o $(GOBIN)/$(BIN_NAME) -v -ldflags "-X main.version=$(VERSION)" .
	@echo "Compilation done."
	@echo "Run \"./build/bin/$(BIN_NAME) -h\" to view available commands."

test:
	go test -v ./...

docker-image:
	cd _assets/docker && $(MAKE) push

build-platforms:
	xgo \
		-ldflags "-X main.version=$(VERSION)" \
		-out $(BIN_NAME) \
		-dest $(GOBIN) \
		-image $(IMAGE_NAME) \
		-targets=$(XGO_TARGETS) .

release:
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is not set. Unable to release to GitHub.)
endif
	# FIXME: this might remove a real release if not careful
	-github-release delete --tag $(VERSION)
	github-release release --tag $(VERSION) --draft
	cd $(GOBIN); \
	for FILE in $$(ls); do \
		github-release upload \
			--tag $(VERSION) \
			--file $${FILE} \
			--name $${FILE} \
			--replace; \
	done

clean:
	rm -f $(GOBIN)/*
