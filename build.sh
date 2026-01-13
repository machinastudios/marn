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

# Create dist directory
mkdir -p dist

# Get dependencies
echo -e "${GREEN}Downloading dependencies...${NC}"
go mod tidy

# Build for Linux amd64
echo -e "${GREEN}Building for Linux (amd64)...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION}" -o dist/marn-linux-amd64 .

# Build for Linux ARM64
echo -e "${GREEN}Building for Linux (arm64)...${NC}"
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=${VERSION}" -o dist/marn-linux-arm64 .

# Build for Windows amd64
echo -e "${GREEN}Building for Windows (amd64)...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION}" -o dist/marn-windows-amd64.exe .

# Build for macOS amd64
echo -e "${GREEN}Building for macOS (amd64)...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION}" -o dist/marn-darwin-amd64 .

# Build for macOS ARM64 (M1/M2)
echo -e "${GREEN}Building for macOS (arm64)...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=${VERSION}" -o dist/marn-darwin-arm64 .

# Build for current platform
echo -e "${GREEN}Building for current platform...${NC}"
go build -ldflags "-X main.Version=${VERSION}" -o dist/marn .

echo ""
echo -e "${GREEN}✓ Build complete!${NC}"
echo ""
echo "Binaries are in the dist/ directory:"
ls -la dist/
