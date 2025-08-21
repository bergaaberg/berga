#!/bin/bash
# Berga CLI Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/bergaaberg/berga/main/install/install.sh | bash

set -e

# Configuration
REPO="bergaaberg/berga"
BINARY_NAME="berga"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_os_arch() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    case $OS in
        Darwin)
            OS="darwin"
            ;;
        Linux)
            OS="linux"
            ;;
        *)
            log_error "Unsupported OS: $OS"
            ;;
    esac
    
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        i386|i686)
            ARCH="386"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            ;;
    esac
    
    log_info "Detected OS: $OS, Architecture: $ARCH"
}

# Get the latest release version
get_latest_version() {
    log_info "Fetching latest release information..."
    
    if command -v curl >/dev/null 2>&1; then
        LATEST_VERSION=$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    else
        log_error "Neither curl nor wget found. Please install one of them."
    fi
    
    if [ -z "$LATEST_VERSION" ]; then
        log_error "Failed to get latest release version"
    fi
    
    log_info "Latest version: $LATEST_VERSION"
}

# Download and install
download_and_install() {
    BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$BINARY_FILE"
    TEMP_FILE="/tmp/$BINARY_FILE"
    
    log_info "Downloading from: $DOWNLOAD_URL"
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL -o "$TEMP_FILE" "$DOWNLOAD_URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "$TEMP_FILE" "$DOWNLOAD_URL"
    else
        log_error "Neither curl nor wget found"
    fi
    
    if [ ! -f "$TEMP_FILE" ]; then
        log_error "Failed to download binary"
    fi
    
    # Make it executable
    chmod +x "$TEMP_FILE"
    
    # Check if we need sudo for installation
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        log_info "Installing to $INSTALL_DIR (requires sudo)"
        sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    log_success "$BINARY_NAME installed to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    if command -v $BINARY_NAME >/dev/null 2>&1; then
        INSTALLED_VERSION=$($BINARY_NAME --version 2>/dev/null | grep -o 'v[0-9.]*' || echo "unknown")
        log_success "Installation verified! Version: $INSTALLED_VERSION"
        log_info "Run '$BINARY_NAME --help' to get started"
        log_info "Run '$BINARY_NAME config init' to initialize your configuration"
    else
        log_warning "Binary installed but not found in PATH. You may need to:"
        log_warning "1. Restart your terminal"
        log_warning "2. Add $INSTALL_DIR to your PATH"
        log_warning "3. Run: export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# Main installation flow
main() {
    log_info "Starting Berga CLI installation..."
    
    detect_os_arch
    get_latest_version
    download_and_install
    verify_installation
    
    log_success "Installation complete!"
    echo ""
    echo "Quick start:"
    echo "  $BINARY_NAME config init    # Initialize configuration"
    echo "  $BINARY_NAME script list    # List available scripts"
    echo "  $BINARY_NAME --help         # Show help"
}

# Run main function
main "$@"
