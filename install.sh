#!/bin/sh
# install.sh — Install propcheck-ai from GitHub Releases
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/mauricioTechDev/propcheck-ai/main/install.sh | sh
#   curl -sSL ... | INSTALL_DIR=/custom/path sh
#   curl -sSL ... | VERSION=0.1.0 sh

set -eu

REPO="mauricioTechDev/propcheck-ai"
BINARY="propcheck-ai"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

fail() {
    printf "Error: %s\n" "$1" >&2
    exit 1
}

info() {
    printf "  %s\n" "$1"
}

detect_os() {
    os="$(uname -s)"
    case "$os" in
        Darwin)  echo "darwin" ;;
        Linux)   echo "linux" ;;
        *)       fail "Unsupported OS: $os. Only macOS and Linux are supported." ;;
    esac
}

detect_arch() {
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)   echo "arm64" ;;
        *)               fail "Unsupported architecture: $arch. Only amd64 and arm64 are supported." ;;
    esac
}

detect_downloader() {
    if command -v curl >/dev/null 2>&1; then
        echo "curl"
    elif command -v wget >/dev/null 2>&1; then
        echo "wget"
    else
        fail "Neither curl nor wget found. Please install one of them."
    fi
}

download() {
    url="$1"
    output="$2"
    case "$(detect_downloader)" in
        curl) curl -fsSL -o "$output" "$url" ;;
        wget) wget -q -O "$output" "$url" ;;
    esac
}

resolve_version() {
    if [ -n "${VERSION:-}" ]; then
        echo "$VERSION"
        return
    fi

    tag=""
    case "$(detect_downloader)" in
        curl) tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')" ;;
        wget) tag="$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')" ;;
    esac

    if [ -z "$tag" ]; then
        fail "Could not determine latest version. Set VERSION explicitly: VERSION=0.1.0 sh install.sh"
    fi
    echo "$tag"
}

verify_checksum() {
    archive_path="$1"
    checksum_path="$2"
    archive_name="$(basename "$archive_path")"

    expected="$(grep "$archive_name" "$checksum_path" | awk '{print $1}')"
    if [ -z "$expected" ]; then
        fail "No checksum found for $archive_name in checksums file."
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        actual="$(sha256sum "$archive_path" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual="$(shasum -a 256 "$archive_path" | awk '{print $1}')"
    else
        printf "  Warning: sha256sum/shasum not found, skipping checksum verification.\n" >&2
        return 0
    fi

    if [ "$expected" != "$actual" ]; then
        fail "Checksum mismatch for $archive_name. Expected: $expected, Actual: $actual"
    fi
}

main() {
    os="$(detect_os)"
    arch="$(detect_arch)"
    version="$(resolve_version)"

    printf "Installing %s v%s (%s/%s)\n" "$BINARY" "$version" "$os" "$arch"

    archive_name="${BINARY}_${version}_${os}_${arch}.tar.gz"
    base_url="https://github.com/${REPO}/releases/download/v${version}"

    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    info "Downloading ${archive_name}..."
    download "${base_url}/${archive_name}" "${tmpdir}/${archive_name}"

    info "Downloading checksums..."
    download "${base_url}/checksums.txt" "${tmpdir}/checksums.txt"

    info "Verifying checksum..."
    verify_checksum "${tmpdir}/${archive_name}" "${tmpdir}/checksums.txt"

    info "Extracting..."
    tar -xzf "${tmpdir}/${archive_name}" -C "${tmpdir}"

    info "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        cp "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
        chmod +x "${INSTALL_DIR}/${BINARY}"
    else
        printf "  Need elevated permissions for %s\n" "$INSTALL_DIR"
        sudo cp "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY}"
    fi

    printf "\nInstalled %s v%s to %s/%s\n" "$BINARY" "$version" "$INSTALL_DIR" "$BINARY"
    printf "Run '%s version' to verify.\n" "$BINARY"
}

main
