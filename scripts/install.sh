#!/bin/sh
# ==============================================================================
#  bgscan installer
#  https://github.com/MohsenBg/bgscan
# ------------------------------------------------------------------------------
#  Installs bgscan with the native Go installer (bgscan-builder). The builder
#  is resolved from PATH or downloaded from its GitHub release, then hands
#  off to `bgscan-builder install`, which resolves the latest (or a pinned)
#  release, verifies its SHA-256 checksum, and installs into ./bgscan.
#
#  This script is fully standalone and safe to pipe:
#
#    curl -fsSL <raw-url> | sh
#    curl -fsSL <raw-url> | sh -s -- --version v2.10.0
#    curl -fsSL <raw-url> | sh -s -- --dir ./bgscan-dev
# ==============================================================================

set -eu

REPOSITORY_OWNER="MohsenBg"
REPOSITORY_NAME="bgscan-builder"
VERSION="latest"
DEST_DIR="${DEST_DIR:-./bgscan}"

usage() {
  echo "Usage: install.sh [--version <tag|latest>] [--dir <path>]" >&2
  echo "" >&2
  echo "Examples:" >&2
  echo "  curl -fsSL <raw-url> | sh" >&2
  echo "  curl -fsSL <raw-url> | sh -s -- --version v2.10.0" >&2
  echo "  curl -fsSL <raw-url> | sh -s -- --dir ./bgscan-dev" >&2
  exit 1
}

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
    ;;
  *)
    usage
    ;;
  esac
done

resolve_builder() {
  cmd="$(command -v bgscan-builder 2>/dev/null || true)"

  if [ -n "$cmd" ]; then
    BUILDER="$cmd"
    return 0
  fi

  BUILDER=""
  return 1
}


download_builder() {
  if [ -n "${TERMUX_VERSION:-}" ] ||
     [ "${PREFIX:-}" = "/data/data/com.termux/files/usr" ]; then

    os="android"

    case "$(uname -m)" in
    aarch64 | arm64)
      arch="arm64-v8a"
      ;;
    armv7l | armv7)
      arch="armeabi-v7a"
      ;;
    x86_64)
      arch="x86_64"
      ;;
    i386 | i686)
      arch="x86"
      ;;
    *)
      echo "error: unsupported Termux architecture: $(uname -m)" >&2
      exit 1
      ;;
    esac

  else


    case "$(uname -s)" in
    Linux*)
      os="linux"
      ;;
    Darwin*)
      os="macos"
      ;;
    *)
      echo "error: unsupported system for the builder: $(uname -s)" >&2
      exit 1
      ;;
    esac

    case "$(uname -m)" in
    x86_64)
      arch="64"
      ;;
    aarch64 | arm64)
      arch="arm64"
      ;;
    armv7l | armv7)
      arch="arm32-v7a"
      ;;
    i386 | i686)
      arch="32"
      ;;
    *)
      echo "error: unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
    esac

  fi

  asset="bgscan-builder-${os}-${arch}"
  url="https://github.com/${REPOSITORY_OWNER}/${REPOSITORY_NAME}/releases/latest/download/${asset}"

  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "$TMP_DIR"' EXIT

  echo "downloading ${REPOSITORY_NAME} (${asset}) ..." >&2

  if ! curl -fsSL --location "$url" -o "$TMP_DIR/bgscan-builder"; then
    echo "error: failed to download ${REPOSITORY_NAME} from ${url}" >&2
    exit 1
  fi

  chmod +x "$TMP_DIR/bgscan-builder"
  BUILDER="$TMP_DIR/bgscan-builder"
}


if ! resolve_builder; then
  echo "bgscan-builder not found on PATH; downloading it ..." >&2
  download_builder
fi

"$BUILDER" install --version "$VERSION" --dir "$DEST_DIR"
