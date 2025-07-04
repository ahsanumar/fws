#!/bin/bash

# FWS (File Watch Server) Installation Script
# 
# This script installs the latest version of fws from GitHub releases
# Usage: curl -fsSL https://raw.githubusercontent.com/ahsanumar/fws/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="ahsanumar/fws"
BINARY_NAME="fws"
INSTALL_DIR="/usr/local/bin"

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
}

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        *)          log_error "Unsupported operating system: $(uname -s)"; exit 1;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64)     arch="amd64";;
        arm64)      arch="arm64";;
        aarch64)    arch="arm64";;
        *)          log_error "Unsupported architecture: $(uname -m)"; exit 1;;
    esac
    
    echo "${os}-${arch}"
}

# Get the latest release version
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    
    echo "$version"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    if ! command_exists curl; then
        log_error "curl is required but not installed."
        exit 1
    fi
    
    if ! command_exists tar; then
        log_error "tar is required but not installed."
        exit 1
    fi
}

# Download and install
install_fws() {
    local platform="$1"
    local version="$2"
    local download_url="https://github.com/${REPO}/releases/download/${version}/fws-${platform}.tar.gz"
    local temp_dir=$(mktemp -d)
    
    log_info "Downloading fws ${version} for ${platform}..."
    
    # Download the archive
    if ! curl -fsSL "$download_url" -o "${temp_dir}/fws.tar.gz"; then
        log_error "Failed to download fws from ${download_url}"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Extract the archive
    log_info "Extracting archive..."
    if ! tar -xzf "${temp_dir}/fws.tar.gz" -C "$temp_dir"; then
        log_error "Failed to extract archive"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Find the binary
    local binary_path="${temp_dir}/fws-${platform}"
    if [ ! -f "$binary_path" ]; then
        log_error "Binary not found in archive"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Install the binary
    log_info "Installing fws to ${INSTALL_DIR}..."
    
    if [ -w "$INSTALL_DIR" ]; then
        cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
}

# Verify installation
verify_installation() {
    if command_exists "$BINARY_NAME"; then
        local installed_version
        installed_version=$("$BINARY_NAME" --help 2>&1 | head -1 || echo "unknown")
        log_success "fws has been installed successfully!"
        log_info "Installed to: $(which $BINARY_NAME)"
        log_info "Try running: $BINARY_NAME --help"
    else
        log_error "Installation verification failed"
        exit 1
    fi
}

# Main installation process
main() {
    echo "======================================"
    echo "FWS (File Watch Server) Installer"
    echo "======================================"
    echo
    
    # Check prerequisites
    check_prerequisites
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get latest version
    local version
    version=$(get_latest_version)
    log_info "Latest version: $version"
    
    # Check if already installed
    if command_exists "$BINARY_NAME"; then
        log_warning "fws is already installed. This will overwrite the existing installation."
        read -p "Continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Installation cancelled."
            exit 0
        fi
    fi
    
    # Install
    install_fws "$platform" "$version"
    
    # Verify
    verify_installation
    
    echo
    log_success "Installation complete!"
    echo
    echo "Quick start:"
    echo "  $BINARY_NAME init                    # Initialize configuration"
    echo "  $BINARY_NAME --mode uploader        # Run uploader mode"
    echo "  $BINARY_NAME --mode watcher --daemon # Run watcher as daemon"
    echo
    echo "For more information: https://github.com/${REPO}"
}

# Handle script interruption
trap 'log_error "Installation interrupted"; exit 1' INT TERM

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            echo "FWS Installation Script"
            echo
            echo "Usage: $0 [options]"
            echo
            echo "Options:"
            echo "  --help, -h     Show this help message"
            echo
            echo "Environment variables:"
            echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
            echo
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
    shift
done

# Override install directory if specified
if [ -n "$INSTALL_DIR_OVERRIDE" ]; then
    INSTALL_DIR="$INSTALL_DIR_OVERRIDE"
fi

# Run main installation
main 