#!/usr/bin/env bash
# ==============================================================================
# release.sh
# ------------------------------------------------------------------------------
# CI RELEASE ORCHESTRATOR (GITHUB ACTIONS ONLY)
#
# FLOW:
#   1. Ensure bgscan-builder exists (via install-builder.sh)
#   2. Validate project root (go.mod module bgscan)
#   3. Run bgscan-builder release command
#   4. If Android build:
#        - install Android NDK
#        - pass --ndk-dir to builder
#   5. Output:
#        dist/*.zip ONLY (no raw binaries kept)
#
# USAGE:
#   ./release.sh <os> <arch> <project-dir>
#
#   os           Target OS (e.g. linux, macos, windows, android). Default: linux
#   arch         Target architecture, or "all". Default: all
#   project-dir  Path to project root. Default: current directory
#
# IMPORTANT:
#   - THIS SCRIPT IS CI ONLY
# ==============================================================================
set -euo pipefail

# ==============================================================================
# INPUTS
# ==============================================================================

OS="${1:-linux}"
ARCH="${2:-all}"
PROJECT_DIR="${3:-$PWD}"
DEST_DIR="$PROJECT_DIR/dist"
ROOT_DIR="$PROJECT_DIR"
INSTALLER="$ROOT_DIR/install-builder.sh"
BUILDER="$ROOT_DIR/bgscan-builder"
SCRIPT_NAME="$(basename "$0")"
START_TIME=$(date +%s)
STEP=0

# ==============================================================================
# LOGGER
# ==============================================================================
# Color codes (disabled automatically if output isn't a terminal, e.g. when
# captured by some CI log viewers — GitHub Actions itself renders ANSI fine)
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
  date -u '+%Y-%m-%d %H:%M:%S UTC'
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
# VALIDATE PROJECT ROOT
# ==============================================================================

validate_project() {
  log "Validating project root"

  info "Project directory : $ROOT_DIR"
  info "Target OS         : $OS"
  info "Target arch       : $ARCH"

  [ -f "$ROOT_DIR/go.mod" ] ||
    fail "go.mod not found in $ROOT_DIR"
  success "go.mod located"

  grep -q "module bgscan" "$ROOT_DIR/go.mod" ||
    fail "Invalid module in go.mod (expected 'bgscan')"
  success "Module name validated"

  info "Creating output directory: $DEST_DIR"
  mkdir -p "$DEST_DIR"
  success "Output directory ready"
}

# ==============================================================================
# INSTALL BUILDER IF NOT EXISTS
# ==============================================================================

ensure_builder() {
  log "Checking for bgscan-builder binary"

  if [ -f "$BUILDER" ]; then
    success "bgscan-builder already present at $BUILDER"
  else
    warn "bgscan-builder not found at $BUILDER"

    [ -f "$INSTALLER" ] ||
      fail "Installer script not found at $INSTALLER"

    info "Running installer: $INSTALLER"
    chmod +x "$INSTALLER"
    "$INSTALLER"

    [ -f "$BUILDER" ] ||
      fail "Installer completed but $BUILDER still not found"

    success "bgscan-builder installed successfully"
  fi

  chmod +x "$BUILDER"
}

# ==============================================================================
# ANDROID NDK SETUP
# ==============================================================================

setup_android_ndk() {
  log "Setting up Android NDK"

  local api="21"
  local ndk_version="r27d"
  NDK_DIR="$ROOT_DIR/android-ndk-${ndk_version}"

  info "Target API level : $api"
  info "NDK version      : $ndk_version"
  info "NDK directory    : $NDK_DIR"

  info "Updating apt package index"
  sudo apt-get update -y >/dev/null ||
    fail "apt-get update failed"

  info "Installing required packages: wget unzip curl build-essential"
  sudo apt-get install -y wget unzip curl build-essential >/dev/null ||
    fail "apt-get install failed"
  success "Build dependencies installed"

  if [ -d "$NDK_DIR" ]; then
    success "NDK already present, skipping download"
  else
    info "Downloading Android NDK ${ndk_version}"
    wget -q \
      "https://dl.google.com/android/repository/android-ndk-${ndk_version}-linux.zip" \
      -O "$ROOT_DIR/ndk.zip" ||
      fail "Failed to download Android NDK ${ndk_version}"
    success "NDK archive downloaded"

    info "Extracting NDK archive to $ROOT_DIR"
    unzip -q "$ROOT_DIR/ndk.zip" -d "$ROOT_DIR" ||
      fail "Failed to extract NDK archive"
    rm -f "$ROOT_DIR/ndk.zip"
    success "NDK extracted and archive cleaned up"
  fi

  export NDK_DIR="$NDK_DIR"
  export TOOLCHAIN="$NDK_DIR/toolchains/llvm/prebuilt/linux-x86_64"
  export PATH="$TOOLCHAIN/bin:$PATH"

  [ -d "$TOOLCHAIN" ] ||
    fail "NDK toolchain directory not found at $TOOLCHAIN — NDK layout may have changed"

  success "Android NDK environment configured"
}

# ==============================================================================
# RELEASE EXECUTION
# ==============================================================================

run_release() {
  log "Running bgscan-builder release"

  ARGS=(release --project-dir "$PROJECT_DIR" --os "$OS" --arch "$ARCH" --dest "$DEST_DIR")

  if [ "$OS" = "android" ]; then
    info "Android target detected — preparing NDK toolchain"
    setup_android_ndk
    ARGS+=(--ndk-dir "$NDK_DIR")
  fi

  info "Executing: $BUILDER ${ARGS[*]}"
  "$BUILDER" "${ARGS[@]}" ||
    fail "bgscan-builder release command failed"

  success "Release build completed"
}

# ==============================================================================
# POST PROCESS: ZIP ONLY (CLEAN OUTPUT)
# ==============================================================================

package_artifacts() {
  log "Packaging artifacts (zip-only output)"

  cd "$DEST_DIR" ||
    fail "Failed to enter dist directory: $DEST_DIR"

  local count=0
  for file in *; do
    [ -f "$file" ] || continue
    local zip_name="${file}.zip"

    info "Compressing: $file -> $zip_name"
    zip -q "$zip_name" "$file" ||
      fail "Failed to zip $file"

    rm -f "$file"
    count=$((count + 1))
  done

  if [ "$count" -eq 0 ]; then
    warn "No raw artifacts found to package in $DEST_DIR"
  else
    success "Packaged ${count} artifact(s) into zip archives"
  fi
}

# ==============================================================================
# MAIN
# ==============================================================================

main() {
  log "Starting BgScan release pipeline"

  validate_project
  ensure_builder
  run_release
  package_artifacts

  local end_time elapsed
  end_time=$(date +%s)
  elapsed=$((end_time - START_TIME))

  echo
  echo -e "${C_GREEN}════════════════════════════════════════════════════════════${C_RESET}"
  echo -e "${C_GREEN}  Release pipeline completed successfully${C_RESET}"
  echo -e "${C_GREEN}  OS          : ${OS}${C_RESET}"
  echo -e "${C_GREEN}  Arch        : ${ARCH}${C_RESET}"
  echo -e "${C_GREEN}  Output dir  : ${DEST_DIR}${C_RESET}"
  echo -e "${C_GREEN}  Total steps : ${STEP}${C_RESET}"
  echo -e "${C_GREEN}  Elapsed time: ${elapsed}s${C_RESET}"
  echo -e "${C_GREEN}  Finished at : $(_timestamp)${C_RESET}"
  echo -e "${C_GREEN}════════════════════════════════════════════════════════════${C_RESET}"
  echo
  echo "Final artifacts:"
  ls -lh "$DEST_DIR"
}

main
