VERSION=$(shell git rev-parse HEAD)
BINARY=add-deploy-imgs-to-related-imgs
RELEASE_TAG ?= "unknown"
# This is a test comment - it will be removed.

# build for your system
.PHONY: build
build:
	go build -o $(BINARY) -ldflags "-X github.com/opdev/add-deploy-imgs-to-related-imgs/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep add-deploy-imgs-to-related-imgs

.PHONY: tidy
tidy:
	go mod tidy -compat=1.17

.PHONY: build-cross
build-cross:
	make build-linux
	make build-darwin

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)_linux_amd64 -ldflags "-X github.com/opdev/add-deploy-imgs-to-related-imgs/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep $(BINARY)_linux_amd64

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY)_darwin_amd64 -ldflags "-X github.com/opdev/add-deploy-imgs-to-related-imgs/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep $(BINARY)_darwin_amd64

.PHONY: test
test: build
	./test/test.sh