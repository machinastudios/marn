#!/bin/bash
# Build script for marn - builds for Windows, Linux, and macOS

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get version from git tag or use dev
VERSION=$(git describe --tags 2>/dev/null || echo "dev-$(git rev-parse --short HEAD 2>/dev/null || echo 'local')")

echo -e "${BLUE}Building marn...${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo ""

# Create dist directories
mkdir -p dist/linux dist/macos

# Change to src directory
cd src

# Get dependencies
echo -e "${GREEN}Downloading dependencies...${NC}"
go mod tidy

# Build for Windows amd64
echo -e "${GREEN}Building for Windows (amd64)...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION}" -o ../dist/marn.exe .

# Build for Linux amd64
echo -e "${GREEN}Building for Linux (amd64)...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION}" -o ../dist/linux/marn .

# Build for macOS amd64
echo -e "${GREEN}Building for macOS (amd64)...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION}" -o ../dist/macos/marn .

cd ..

echo ""
echo -e "${GREEN}✓ Build complete!${NC}"
echo ""
echo "Binaries are in the dist/ directory:"
ls -la dist/
