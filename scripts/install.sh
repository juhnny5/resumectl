#!/bin/bash
# install.sh - Installation script for resumectl

set -e

REPO="resumectl"
BINARY_NAME="resumectl"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Detected: $OS-$ARCH"

# Build from source
echo "Building from source..."
go build -o "$BINARY_NAME" ./cmd/resumectl

# Install
echo "Installing to $INSTALL_DIR..."
sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "Successfully installed $BINARY_NAME to $INSTALL_DIR"
echo ""
echo "Run 'resumectl --help' to get started"
