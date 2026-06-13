#!/usr/bin/env bash
# IconHash installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/install.sh | bash
#
# This script detects your OS and architecture, downloads the latest
# IconHash binary from GitHub Releases, and installs it to /usr/local/bin.

set -e

REPO="cyberspacesec/iconhash-skills"
BINARY="iconhash"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${CYAN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        FreeBSD*) echo "freebsd" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *) error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "x86_64" ;;
        aarch64|arm64) echo "aarch64" ;;
        i386|i686)     echo "i386" ;;
        armv6l|armv7l) echo "armv6h" ;;
        riscv64)       echo "riscv64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest version tag from GitHub
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || echo "latest"
}

# Main
main() {
    echo ""
    echo "  ██▓ ▄████▄   ▒█████   ███▄    █     ██░ ██  ▄▄▄       ██████  ██░ ██"
    echo " ▓██▒▒██▀ ▀█  ▒██▒  ██▒ ██ ▀█   █    ▓██░ ██▒▒████▄   ▒██    ▒ ▓██░ ██▒"
    echo " ▒██▒▒▓█    ▄ ▒██░  ██▒▓██  ▀█ ██▒   ▒██▀▀██░▒██  ▀█▄ ░ ▓██▄   ▒██▀▀██░"
    echo " ░██░▒▓▓▄ ▄██▒▒██   ██░▓██▒  ▐▌██▒   ░▓█ ░██ ░██▄▄▄▄██  ▒   ██▒░▓█ ░██"
    echo " ░██░▒ ▓███▀ ░░ ████▓▒░▒██░   ▓██░   ░▓█▒░██▓ ▓█   ▓██▒▒██████▒▒░▓█▒░██▓"
    echo ""
    info "IconHash Installer"
    echo ""

    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)

    info "Platform: ${OS}/${ARCH}"
    info "Version:  ${VERSION}"
    echo ""

    # Construct download URL
    if [ "$OS" = "windows" ]; then
        EXT="zip"
        ARCHIVE="iconhash_lite_${OS}_${ARCH}.${EXT}"
    else
        EXT="tar.gz"
        ARCHIVE="iconhash_lite_${OS}_${ARCH}.${EXT}"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ARCHIVE}"

    # Create temp directory
    TMPDIR=$(mktemp -d)
    trap "rm -rf $TMPDIR" EXIT

    # Download
    info "Downloading ${ARCHIVE}..."
    if command -v wget &>/dev/null; then
        wget -q --show-progress -O "${TMPDIR}/${ARCHIVE}" "${DOWNLOAD_URL}" || \
            error "Download failed. URL: ${DOWNLOAD_URL}"
    elif command -v curl &>/dev/null; then
        curl -fSL --progress-bar -o "${TMPDIR}/${ARCHIVE}" "${DOWNLOAD_URL}" || \
            error "Download failed. URL: ${DOWNLOAD_URL}"
    else
        error "Neither wget nor curl found. Please install one and retry."
    fi

    # Extract
    info "Extracting..."
    cd "$TMPDIR"
    if [ "$EXT" = "zip" ]; then
        unzip -q "${ARCHIVE}" || error "Extraction failed"
    else
        tar xzf "${ARCHIVE}" || error "Extraction failed"
    fi

    # Find the binary
    if [ -f "${BINARY}" ]; then
        :
    elif [ -f "${BINARY}.exe" ]; then
        BINARY="${BINARY}.exe"
    else
        error "Binary not found in archive"
    fi

    # Install
    chmod +x "${BINARY}"

    if [ -w "${INSTALL_DIR}" ]; then
        mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        warn "No write permission to ${INSTALL_DIR}. Using sudo..."
        sudo mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi

    # Verify
    if command -v iconhash &>/dev/null; then
        success "Installed: ${INSTALL_DIR}/${BINARY}"
        echo ""
        iconhash --version
        echo ""
        info "Quick start:"
        echo "  iconhash lookup -- -305179312     # Lookup a fingerprint"
        echo "  iconhash identify https://site    # Identify a website"
        echo "  iconhash fingerprints update      # Update fingerprint database"
    else
        warn "Binary installed but not in PATH."
        info "Add ${INSTALL_DIR} to your PATH or run with full path."
    fi
}

main "$@"
