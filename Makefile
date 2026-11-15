APP_NAME ?= go-product
BUILD_DIR ?= bin
MAIN_SRC ?= ./cmd/server
DOCKER_IMAGE ?= $(APP_NAME):latest
DOCKERFILE ?= deploy/Dockerfile

.PHONY: all build run docker-build clean test help

all: build

## run: Run application locally with go run
run:
	go run $(MAIN_SRC)

## build: Build binary locally into bin/
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_SRC)

## docker-build: Build Docker image using deploy/Dockerfile
docker-build:
	docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE) .

## test: Run tests
test:
	go test -v ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
