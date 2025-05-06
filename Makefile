# Define the binary names
API_BINARY=nba-salaries-api
DATA_BINARY=nba-salaries-data
# Define the build output directory
BUILD_DIR=bin
# Define the source directory
SRC_DIR=cmd
# Default target - builds both binaries
all: build-api build-data
# Build the API application
build-api:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(API_BINARY) ./cmd/api
# Build the data application
build-data:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(DATA_BINARY) ./cmd/data
# Combined build target
build: build-api build-data
# Run the API application
run-api: build-api
	$(BUILD_DIR)/$(API_BINARY)
# Run the data application
run-data: build-data
	$(BUILD_DIR)/$(DATA_BINARY)
# Clean the build output
clean:
	rm -rf $(BUILD_DIR)
# Test the application
test:
	go test ./...
# Format the code
fmt:
	go fmt ./...
# Run data script directly without building
data:
	go run cmd/data/main.go
# Run the API server directly without building
api:
	go run cmd/api/main.go
# List all available targets
.PHONY: all build build-api build-data run-api run-data clean test fmt data api help
help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  build:      Build both API and data applications"
	@echo "  build-api:  Build just the API application"
	@echo "  build-data: Build just the data application"
	@echo "  run-api:    Build and run the API application"
	@echo "  run-data:   Build and run the data application"
	@echo "  clean:      Clean the build output"
	@echo "  test:       Run the tests"
	@echo "  fmt:        Format the code"
	@echo "  data:       Run data script directly (without building)"
	@echo "  api:        Run API server directly (without building)"
	@echo "  help:       Show this help message"
