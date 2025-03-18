# Variables
APP_NAME := whatsmyip
BUILD_DIR := build
DOCKER_IMAGE := whatsmyip
DOCKER_TAG := latest

# Default target
.PHONY: all
all: build

# Build the application binary
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(APP_NAME) .

# Run the application
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	@go run .

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Build Docker image
.PHONY: docker
docker:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# Help command
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build   - Build the $(APP_NAME) binary"
	@echo "  make run     - Run the application with go run"
	@echo "  make test    - Run tests"
	@echo "  make docker  - Build Docker image"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make help    - Show this help message"
