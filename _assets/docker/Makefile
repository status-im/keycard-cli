GIT_COMMIT = $(shell git rev-parse --short HEAD)
GIT_ROOT = $(shell git rev-parse --show-toplevel)

IMAGE_TAG  ?= latest
IMAGE_NAME ?= statusteam/keycard-cli-ci:$(IMAGE_TAG)

build:
	docker build \
		--label "commit=$(GIT_COMMIT)" \
		-t $(IMAGE_NAME) .

push: build
	docker push $(IMAGE_NAME)

clean:
	docker rmi -f $(IMAGE_NAME)
