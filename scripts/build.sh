#!/bin/bash

# Build script for wanikani-api

set -e

BINARY_NAME="wanikani-api"
BUILD_DIR="bin"
CMD_PATH="./cmd/wanikani-api"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Building ${BINARY_NAME}...${NC}"

# Create build directory
mkdir -p "$BUILD_DIR"

# Build the binary
go build -o "$BUILD_DIR/$BINARY_NAME" "$CMD_PATH"

echo -e "${GREEN}✓ Build complete: $BUILD_DIR/$BINARY_NAME${NC}"
echo ""
echo "To run the application:"
echo "  ./$BUILD_DIR/$BINARY_NAME"
