# Go parameters
BINARY_NAME=gotris
MAIN_PACKAGE=./cmd/main.go
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOMOD=$(GO) mod

# Build directory
BUILD_DIR=build
# Linting and formatting
GOLINT=golangci-lint

.PHONY: all build test clean run deps lint fmt tidy help

all: test build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PACKAGE)

test:
	$(GOTEST) ./... -v

vet:
	$(GOVET) ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

deps:
	$(GOGET) -u ./...

lint:
	$(GOLINT) run

fmt:
	$(GO) fmt ./...

tidy:
	$(GOMOD) tidy

help:
	@echo "make - Build and test the application"
	@echo "make build - Build the application"
	@echo "make build-linux - Build for Linux platform"
	@echo "make test - Run tests"
	@echo "make vet - Run go vet"
	@echo "make clean - Clean build files"
	@echo "make run - Build and run the application"
	@echo "make deps - Update dependencies"
	@echo "make lint - Run linter"
	@echo "make fmt - Format Go code"
	@echo "make tidy - Tidy go.mod file"
	@echo "make help - Display this help"
