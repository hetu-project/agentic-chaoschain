# Go project build configuration
APP_NAME = samples
BUILD_DIR = build
SRCS = $(wildcard *.go) $(wildcard */*.go)

all: deps build

build: copy-config
	@echo "Building Go project..."
	@go build -ldflags "-X main.GitCommit=$(shell git rev-parse HEAD)" -o $(BUILD_DIR)/$(APP_NAME)

copy-config:
	@echo "Copying config files..."
	@mkdir -p $(BUILD_DIR)/config
	@cp -r config/*.toml $(BUILD_DIR)/config/

run-service: build
	@echo "Starting agent service..."
	@./$(BUILD_DIR)/$(APP_NAME) start

register-agent: build
	@echo "Registering agent..."
	@./$(BUILD_DIR)/$(APP_NAME) register --name "$(NAME)" --url "$(URL)" --intro "$(INTRO)"


propose: build
	@echo "Submitting proposal..."
	@./$(BUILD_DIR)/$(APP_NAME) propose  --title "$(TITLE)" --data "$(DATA)"


clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	
.PHONY: all build run deps test clean fmt

help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run-service   - Start the agent service"
	@echo "  make register-agent  NAME=... URL=... INTRO=... - Register an agent"
	@echo "  make propose TITLE=... DATA=... - Submit a proposal"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make help          - Show this help"