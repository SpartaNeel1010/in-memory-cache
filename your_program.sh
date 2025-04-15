#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Define app directory and binary name
APP_DIR="./app"
BINARY_NAME="app_binary"

# Build the Go app
echo "Building Go application..."
go build -o "$BINARY_NAME" "$APP_DIR"

# Run the binary with arguments passed to the script
echo "Running application with arguments: $@"
./"$BINARY_NAME" "$@"