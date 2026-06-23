#!/usr/bin/env bash
# ==============================================================================
# bgscan-builder installer
# ------------------------------------------------------------------------------
# Project:
#   bgscan-builder
#
# Purpose:
#   Downloads the correct bgscan-builder binary for the current operating
#   system and CPU architecture, then installs it into the root directory
#   of a valid BgScan project.
#
# Supported Platforms:
#   - Linux AMD64
#   - Linux ARM64
#   - Linux ARMv7
#   - Linux x86
#   - macOS Intel (AMD64)
#   - macOS Apple Silicon (ARM64)
#
# Installation Target:
#   ./bgscan-builder
#
# Safety Checks:
#   - Must be executed from a BgScan project root
#   - Requires go.mod to exist
#   - Requires module name to contain "bgscan"
#
# Example:
#   curl -fsSL <installer-url> | bash
#
#   or
#
#   ./install-builder.sh
#
# Notes:
#   - Windows is intentionally excluded because this installer is a shell
#     script and is intended for Unix-like environments.
#   - Windows support should be provided through a dedicated PowerShell
#     installer (install.ps1).
# ==============================================================================

set -euo pipefail

# ==============================================================================
# CONFIGURATION
# ==============================================================================

REPOSITORY_OWNER="MohsenBg"
REPOSITORY_NAME="bgscan-builder"

INSTALL_NAME="bgscan-builder"
SCRIPT_NAME="$(basename "$0")"
START_TIME=$(date +%s)
STEP=0

# ==============================================================================
# LOGGER
# ==============================================================================
# Color codes (disabled automatically if output isn't a terminal)
if [ -t 1 ]; then
  C_RESET='\033[0m'
  C_BLUE='\033[1;34m'
  C_GREEN='\033[1;32m'
  C_YELLOW='\033[1;33m'
  C_RED='\033[1;31m'
  C_DIM='\033[2m'
else
  C_RESET='' C_BLUE='' C_GREEN='' C_YELLOW='' C_RED='' C_DIM=''
fi

_timestamp() {
  date '+%Y-%m-%d %H:%M:%S'
}

log() {
  STEP=$((STEP + 1))
  echo
  echo -e "${C_BLUE}────────────────────────────────────────────────────────────${C_RESET}"
  echo -e "${C_BLUE}[STEP ${STEP}]${C_RESET} ${C_DIM}$(_timestamp)${C_RESET}  ${SCRIPT_NAME}"
  echo -e "${C_BLUE}────────────────────────────────────────────────────────────${C_RESET}"
  echo -e "  $*"
}

info() {
  echo -e "  ${C_DIM}$(_timestamp)${C_RESET} ${C_GREEN}[INFO]${C_RESET}  $*"
}

warn() {
  echo -e "  ${C_DIM}$(_timestamp)${C_RESET} ${C_YELLOW}[WARN]${C_RESET}  $*"
}

success() {
  echo -e "  ${C_DIM}$(_timestamp)${C_RESET} ${C_GREEN}[ OK ]${C_RESET}  $*"
}

fail() {
  echo
  echo -e "${C_RED}────────────────────────────────────────────────────────────${C_RESET}"
  echo -e "${C_RED}[FATAL]${C_RESET} $(_timestamp)  $*"
  echo -e "${C_RED}────────────────────────────────────────────────────────────${C_RESET}" >&2
  echo
  exit 1
}

# ==============================================================================
# PROJECT VALIDATION
# ==============================================================================

validate_project() {
  log "Validating BgScan project"

  info "Checking for go.mod in current directory"
  [ -f "go.mod" ] ||
    fail "go.mod not found. Run this installer from the project root."
  success "go.mod located"

  local module_name
  module_name="$(awk '/^module / {print $2}' go.mod)"

  [ -n "$module_name" ] ||
    fail "Unable to determine Go module name from go.mod"
  info "Module name resolved: ${module_name}"

  if [[ "$module_name" != *bgscan* ]]; then
    fail "Unsupported module '${module_name}'. Expected a BgScan project."
  fi

  success "Project validated — module: ${module_name}"
}

# ==============================================================================
# PLATFORM DETECTION
# ==============================================================================

detect_platform() {
  log "Detecting operating system and architecture"

  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  info "Raw uname values — OS: ${os}, Arch: ${arch}"

  case "$os" in
  Linux)
    PLATFORM="linux"
    ;;
  Darwin)
    PLATFORM="macos"
    ;;
  *)
    fail "Unsupported operating system: ${os}"
    ;;
  esac

  case "$arch" in
  x86_64)
    ARCH="64"
    ;;
  aarch64 | arm64)
    ARCH="arm64"
    ;;
  armv7l | armv7)
    ARCH="arm32-v7a"
    ;;
  i386 | i686)
    ARCH="32"
    ;;
  *)
    fail "Unsupported architecture: ${arch}"
    ;;
  esac

  ASSET_NAME="bgscan-builder-${PLATFORM}-${ARCH}"

  success "Platform resolved: ${PLATFORM} (${arch})"
  info "Target asset name: ${ASSET_NAME}"
}

# ==============================================================================
# RELEASE URL RESOLUTION
# ==============================================================================

build_download_url() {
  log "Resolving release download URL"

  DOWNLOAD_URL="https://github.com/${REPOSITORY_OWNER}/${REPOSITORY_NAME}/releases/latest/download/${ASSET_NAME}"

  info "Repository : ${REPOSITORY_OWNER}/${REPOSITORY_NAME}"
  info "Asset      : ${ASSET_NAME}"
  success "Download URL resolved: ${DOWNLOAD_URL}"
}

# ==============================================================================
# INSTALLATION
# ==============================================================================

install_builder() {
  log "Downloading bgscan-builder binary"

  info "Fetching from: ${DOWNLOAD_URL}"

  curl \
    --fail \
    --silent \
    --show-error \
    --location \
    "$DOWNLOAD_URL" \
    --output "$INSTALL_NAME" ||
    fail "Download failed. Check your network connection and that a release exists for ${ASSET_NAME}."

  success "Binary downloaded to ./${INSTALL_NAME}"

  info "Setting execute permission on ./${INSTALL_NAME}"
  chmod +x "$INSTALL_NAME"

  success "Installation step completed"
}

# ==============================================================================
# POST-INSTALL VERIFICATION
# ==============================================================================

verify_installation() {
  log "Verifying installation"

  [ -f "$INSTALL_NAME" ] ||
    fail "Installation verification failed — ${INSTALL_NAME} not found."

  chmod +x "$INSTALL_NAME"

  local file_info
  file_info="$(ls -lh "$INSTALL_NAME" | awk '{print $5, $NF}')"

  success "Binary verified: ${INSTALL_NAME}"
  info "Details: ${file_info}"
}

# ==============================================================================
# MAIN
# ==============================================================================

main() {
  log "Starting bgscan-builder installation"

  validate_project
  detect_platform
  build_download_url
  install_builder
  verify_installation

  local end_time elapsed
  end_time=$(date +%s)
  elapsed=$((end_time - START_TIME))

  echo
  echo -e "${C_GREEN}════════════════════════════════════════════════════════════${C_RESET}"
  echo -e "${C_GREEN}  bgscan-builder installed successfully${C_RESET}"
  echo -e "${C_GREEN}  Binary      : ./${INSTALL_NAME}${C_RESET}"
  echo -e "${C_GREEN}  Total steps : ${STEP}${C_RESET}"
  echo -e "${C_GREEN}  Elapsed time: ${elapsed}s${C_RESET}"
  echo -e "${C_GREEN}  Finished at : $(_timestamp)${C_RESET}"
  echo -e "${C_GREEN}════════════════════════════════════════════════════════════${C_RESET}"
  echo
}

main "$@"
