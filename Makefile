SHELL := /bin/bash

GO      ?= go
BINARY  ?= pacman
PKG     ?= .
BUILD_DIR := bin

.PHONY: help deps fmt vet build run clean release build-linux build-darwin build-windows test

default: help

help:
	@echo "Available targets:"
	@echo "  deps     - Download/resolve dependencies (go mod tidy)"
	@echo "  fmt      - Format code (go fmt)"
	@echo "  vet      - Run go vet"
	@echo "  build    - Build local binary into $(BUILD_DIR)/$(BINARY)"
	@echo "  run      - Build and run the game"
	@echo "  clean    - Remove build artifacts"
	@echo "  release  - Cross-compile for linux, darwin, windows"
	@echo "  test     - Run unit tests"

deps:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

build: deps fmt vet
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY) $(PKG)

test: deps
	$(GO) test ./...

run: build
	./$(BUILD_DIR)/$(BINARY)

clean:
	rm -rf $(BUILD_DIR)

release: build-linux build-darwin build-windows

build-linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY)-linux-amd64 $(PKG)

build-darwin:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 $(PKG)

build-windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe $(PKG)


