#!/bin/bash
#
# Quick install script for llc
# Downloads and installs the latest release from GitHub
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/kkw-h/llc/main/install.sh | bash
#   # or
#   wget -qO- https://raw.githubusercontent.com/kkw-h/llc/main/install.sh | bash
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="kkw-h/llc"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="llc"

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*|MSYS*) os="windows";;
        *)          os="unknown";;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        armv7l)         arch="arm";;
        i386|i686)      arch="386";;
        *)              arch="unknown";;
    esac

    echo "${os}-${arch}"
}

# Print message functions
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download file
download() {
    local url="$1"
    local output="$2"

    if command_exists curl; then
        curl -fsSL "$url" -o "$output"
    elif command_exists wget; then
        wget -q "$url" -O "$output"
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# Get latest release version
get_latest_version() {
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"

    if command_exists curl; then
        curl -fsSL "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command_exists wget; then
        wget -qO- "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    fi
}

# Main installation
main() {
    echo ""
    echo "═══════════════════════════════════════════════════"
    echo "  llc - Enhanced ls command with file comments"
    echo "═══════════════════════════════════════════════════"
    echo ""

    # Detect platform
    print_info "Detecting platform..."
    PLATFORM=$(detect_platform)
    OS=$(echo "$PLATFORM" | cut -d'-' -f1)
    ARCH=$(echo "$PLATFORM" | cut -d'-' -f2)

    print_info "Detected: $OS ($ARCH)"

    # Check if platform is supported
    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        print_error "Unsupported platform: $PLATFORM"
        print_error "Supported platforms: linux (amd64, arm64), macOS (amd64, arm64)"
        exit 1
    fi

    if [ "$OS" = "windows" ]; then
        print_error "Windows is not supported by this install script."
        print_info "Please download the binary manually from:"
        print_info "  https://github.com/${REPO}/releases/latest"
        exit 1
    fi

    # Get latest version
    print_info "Fetching latest version..."
    VERSION=$(get_latest_version)

    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest version. Please check your internet connection."
        exit 1
    fi

    print_success "Latest version: $VERSION"

    # Construct download URL
    FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${FILENAME}"

    # Download binary
    print_info "Downloading ${FILENAME}..."
    if download "$DOWNLOAD_URL" "$TMP_FILE"; then
        print_success "Download completed"
    else
        print_error "Failed to download binary"
        print_error "URL: $DOWNLOAD_URL"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Make binary executable
    chmod +x "$TMP_FILE"

    # Verify binary works
    print_info "Verifying binary..."
    if "$TMP_FILE" --version >/dev/null 2>&1; then
        VERSION_OUTPUT=$("$TMP_FILE" --version)
        print_success "Binary verified: $VERSION_OUTPUT"
    else
        print_error "Downloaded binary is not valid"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        SUDO=""
    else
        print_warning "Installation requires sudo privileges"
        SUDO="sudo"
    fi

    # Install binary
    print_info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    if $SUDO mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"; then
        print_success "Installation completed"
    else
        print_error "Failed to install binary to ${INSTALL_DIR}"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Cleanup
    rm -rf "$TMP_DIR"

    # Verify installation
    print_info "Verifying installation..."
    if command_exists "$BINARY_NAME"; then
        INSTALLED_VERSION=$($BINARY_NAME --version)
        print_success "llc is now installed!"
        echo ""
        echo "═══════════════════════════════════════════════════"
        echo "  $INSTALLED_VERSION"
        echo "═══════════════════════════════════════════════════"
        echo ""
        echo "Usage examples:"
        echo "  llc              # List current directory"
        echo "  llc -a           # Show all files"
        echo "  llc -h           # Human-readable sizes"
        echo "  llc --help       # Show all options"
        echo ""
    else
        print_warning "Installation completed but ${BINARY_NAME} is not in PATH"
        print_info "You may need to add ${INSTALL_DIR} to your PATH:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

# Run main function
main "$@"
