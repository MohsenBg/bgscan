#!/bin/sh
# ==============================================================================
#  bgscan installer
#  https://github.com/MohsenBg/bgscan
# ------------------------------------------------------------------------------
#  Installs bgscan via the native Go installer (bgscan-builder). The builder
#  is resolved from PATH or downloaded from its GitHub release, then delegates
#  to `bgscan-builder install`, which resolves the latest (or a pinned)
#  release, verifies its SHA-256 checksum, and installs into DEST_DIR.
#
#  Usage (pipe-safe):
#    curl -fsSL <raw-url> | sh
#    curl -fsSL <raw-url> | sh -s -- --version v2.10.0
#    curl -fsSL <raw-url> | sh -s -- --dir ./bgscan-dev
# ==============================================================================
set -eu

# Defaults
REPOSITORY_OWNER="MohsenBg"
REPOSITORY_NAME="bgscan-builder"
VERSION="latest"
DEST_DIR="${DEST_DIR:-./bgscan}"
BUILDER=""
TMP_DIR=""

# Cleanup
cleanup() {
    [ -z "$TMP_DIR" ] || rm -rf "$TMP_DIR"
}
trap cleanup EXIT

# Helpers
die() {
    echo "error: $*" >&2
    exit 1
}

usage() {
    cat >&2 <<EOF
Usage: install.sh [--version <tag|latest>] [--dir <path>]

Options:
  --version <tag|latest>   bgscan version to install (default: latest)
  --dir     <path>         installation directory    (default: ./bgscan)

Examples:
  curl -fsSL <raw-url> | sh
  curl -fsSL <raw-url> | sh -s -- --version v2.10.0
  curl -fsSL <raw-url> | sh -s -- --dir ./bgscan-dev
EOF
    exit 1
}

# Argument parsing
while [ $# -gt 0 ]; do
    case "$1" in
        --version)
            VERSION="${2:-}"
            [ -n "$VERSION" ] || usage
            shift 2
            ;;
        --dir)
            DEST_DIR="${2:-}"
            [ -n "$DEST_DIR" ] || usage
            shift 2
            ;;
        --)
            shift
            break
            ;;
        *)
            usage
            ;;
    esac
done

# Platform detection
detect_platform() {
    # Termux / Android
    if [ -n "${TERMUX_VERSION:-}" ] ||
       [ "${PREFIX:-}" = "/data/data/com.termux/files/usr" ]; then
        OS="android"
        case "$(uname -m)" in
            aarch64|arm64)  ARCH="arm64-v8a"   ;;
            armv7l|armv7)   ARCH="armeabi-v7a" ;;
            x86_64)         ARCH="x86_64"      ;;
            i386|i686)      ARCH="x86"         ;;
            *)              die "unsupported Termux architecture: $(uname -m)" ;;
        esac
        return
    fi

    # Standard POSIX
    case "$(uname -s)" in
        Linux*)  OS="linux" ;;
        Darwin*) OS="macos" ;;
        *)       die "unsupported operating system: $(uname -s)" ;;
    esac

    case "$(uname -m)" in
        x86_64)        ARCH="64"       ;;
        aarch64|arm64) ARCH="arm64"    ;;
        armv7l|armv7)  ARCH="arm32-v7a";;
        i386|i686)     ARCH="32"       ;;
        *)             die "unsupported architecture: $(uname -m)" ;;
    esac
}

# Resolve or download bgscan-builder
resolve_builder() {
    cmd="$(command -v bgscan-builder 2>/dev/null || true)"
    if [ -n "$cmd" ]; then
        BUILDER="$cmd"
        return 0
    fi
    return 1
}

download_builder() {
    detect_platform

    ASSET="bgscan-builder-${OS}-${ARCH}"
    URL="https://github.com/${REPOSITORY_OWNER}/${REPOSITORY_NAME}/releases/latest/download/${ASSET}"

    TMP_DIR="$(mktemp -d)"
    BUILDER="${TMP_DIR}/bgscan-builder"

    echo "bgscan-builder not found on PATH; downloading ${ASSET} ..." >&2

    curl -fsSL --location "$URL" -o "$BUILDER" \
        || die "failed to download ${REPOSITORY_NAME} from ${URL}"

    chmod +x "$BUILDER"
}

# Main
if ! resolve_builder; then
    download_builder
fi

if [ -r /dev/tty ]; then
    "$BUILDER" install --version "$VERSION" --dir "$DEST_DIR" < /dev/tty
else
    "$BUILDER" install --version "$VERSION" --dir "$DEST_DIR"
fi
