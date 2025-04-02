# Makefile for heart-go-api project

# Variables
APP_NAME=auto_verse
MODULES_DIR=Modules
BIN_DIR=bin

# Default target
all: run

# Run the application
run:
	@echo "Starting the application..."
	go run cmd/main.go

# Run with Air (for live reload)
air:
	air -c air.toml

# Build the application
build:
	@echo "Building the application..."
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/main.go
	@echo "Build complete. Executable: $(BIN_DIR)/$(APP_NAME)"
