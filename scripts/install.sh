#!/bin/bash
# Installation script for docgen-tool

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="kyivinua/docgen-tool"
BINARY_NAME="docgen"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo -e "${GREEN}Installing docgen-tool...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.22 or later from https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.22"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo -e "${YELLOW}Warning: Go version $GO_VERSION is older than recommended $REQUIRED_VERSION${NC}"
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo -e "${GREEN}Downloading docgen-tool...${NC}"

# Clone repository
git clone "https://github.com/${REPO}.git" .

echo -e "${GREEN}Building docgen-tool...${NC}"

# Build binary
go build -o "$BINARY_NAME" ./cmd/docgen

# Install binary
echo -e "${GREEN}Installing to ${INSTALL_DIR}...${NC}"

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
else
    echo -e "${YELLOW}Need sudo permissions to install to ${INSTALL_DIR}${NC}"
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

# Make executable
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Clean up
cd -
rm -rf "$TMP_DIR"

echo -e "${GREEN}✓ Installation complete!${NC}"
echo ""
echo "Verify installation:"
echo "  $ docgen version"
echo ""
echo "Get started:"
echo "  $ docgen init"
echo "  $ docgen generate --config ./docgen.yaml"
echo ""
echo "For more information, visit:"
echo "  https://github.com/${REPO}"
