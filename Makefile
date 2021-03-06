REPO := chinnareddy578/kubernetes-github-authn:1.0

# My dockerhub user is 'xjjo'
IMAGE_NAME := $(REPO)
GO_SRC_PATH := /go/src/github.com/$(REPO)
PORT := 3000

ifeq (1,${WITH_DOCKER})
DOCKER_RUN := docker run --rm -i \
	-v `pwd`:$(GO_SRC_PATH) \
	-w $(GO_SRC_PATH)
GO_RUN := $(DOCKER_RUN) golang:1.13-alpine
endif

.PHONY: build
build:
	$(GO_RUN) go build -o _output/main main.go

.PHONY: vendor
vendor:
	$(GLIDE_RUN) glide install

.PHONY: clean
clean:
	rm -rf _output

.PHONY: docker-build
docker-build:
	#WITH_DOCKER=1 make build
	docker build -t $(IMAGE_NAME) .

.PHONY: docker-build-push
docker-build-push:docker-build
	docker push $(IMAGE_NAME)

.PHONY: docker-run
docker-run:
	docker run -it --rm -p $(PORT):3000 $(IMAGE_NAME)