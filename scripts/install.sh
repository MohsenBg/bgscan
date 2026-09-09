#!/bin/sh
# ==============================================================================
#  bgscan installer
#  https://github.com/MohsenBg/bgscan
# ------------------------------------------------------------------------------
#  Installs bgscan via the native rust installer (bgscan-installer). The builder
#  is resolved from PATH or downloaded from its GitHub release, then delegates
#  to `bgscan-installer install`, which resolves the latest (or a pinned)
#  release, verifies its SHA-256 checksum.
#
#  Usage (pipe-safe):
#    curl -fsSL <raw-url> | sh
#    curl -fsSL <raw-url> | sh -s -- --version v2.10.0
# ==============================================================================
set -eu

# Defaults
REPOSITORY_OWNER="MohsenBg"
REPOSITORY_NAME="bgscan-installer"
VERSION="latest"
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
Usage: install.sh [--version <tag|latest>]

Options:
  --version <tag|latest>   bgscan version to install (default: latest)

Examples:
  curl -fsSL <raw-url> | sh
  curl -fsSL <raw-url> | sh -s -- --version v2.10.0
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
    aarch64 | arm64) ARCH="arm64-v8a" ;;
    armv7l | armv7) ARCH="armeabi-v7a" ;;
    x86_64) ARCH="x86_64" ;;
    i386 | i686) ARCH="x86" ;;
    *) die "unsupported Termux architecture: $(uname -m)" ;;
    esac
    return
  fi

  # Standard POSIX
  case "$(uname -s)" in
  Linux*) OS="linux" ;;
  Darwin*) OS="macos" ;;
  *) die "unsupported operating system: $(uname -s)" ;;
  esac

  case "$(uname -m)" in
  x86_64) ARCH="64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  armv7l | armv7) ARCH="arm32-v7a" ;;
  i386 | i686) ARCH="32" ;;
  *) die "unsupported architecture: $(uname -m)" ;;
  esac
}

# Resolve or download bgscan-installer
resolve_builder() {
  cmd="$(command -v bgscan-installer 2>/dev/null || true)"
  if [ -n "$cmd" ]; then
    BUILDER="$cmd"
    return 0
  fi
  return 1
}

download_builder() {
  detect_platform

  ASSET="bgscan-installer-${OS}-${ARCH}"
  URL="https://github.com/${REPOSITORY_OWNER}/${REPOSITORY_NAME}/releases/latest/download/${ASSET}"

  TMP_DIR="$(mktemp -d)"
  BUILDER="${TMP_DIR}/bgscan-installer"

  echo "bgscan-installer not found on PATH; downloading ${ASSET} ..." >&2

  curl -fsSL --location "$URL" -o "$BUILDER" ||
    die "failed to download ${REPOSITORY_NAME} from ${URL}"

  chmod +x "$BUILDER"
}

# Main
if ! resolve_builder; then
  download_builder
fi

if [ -r /dev/tty ]; then
  "$BUILDER" install --version "$VERSION" </dev/tty
else
  "$BUILDER" install --version "$VERSION"
fi
