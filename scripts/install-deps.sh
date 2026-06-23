#!/usr/bin/env bash
# ==============================================================================
# bgscan dependency installer
# ------------------------------------------------------------------------------
# Project:
#   BgScan
#
# Purpose:
#   Ensures bgscan-builder is installed and executes the dependency bootstrap
#   process required for BgScan development.
#
# Workflow:
#   1. Validate current directory is a BgScan project.
#   2. Install bgscan-builder if missing.
#   3. Ensure executable permissions.
#   4. Execute:
#
#        bgscan-builder setup-dev --project-dir <project-root>
#
# Notes:
#   - Must be executed from the root of a BgScan project.
#   - Requires go.mod to be present.
#   - Intended for developers and CI environments.
# ==============================================================================
set -euo pipefail

PROJECT_ROOT="$PWD"
BUILDER_BINARY="./bgscan-builder"
INSTALL_SCRIPT="./scripts/install-builder.sh"
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
  log "Validating BgScan project structure"

  info "Checking for go.mod in: ${PROJECT_ROOT}"
  [ -f "$PROJECT_ROOT/go.mod" ] ||
    fail "go.mod not found in ${PROJECT_ROOT}. Are you in the project root?"
  success "go.mod located"

  local module_name
  module_name="$(awk '/^module / {print $2}' "$PROJECT_ROOT/go.mod")"

  [ -n "$module_name" ] ||
    fail "Unable to determine module name from go.mod"
  info "Module name resolved: ${module_name}"

  if [[ "$module_name" != *bgscan* ]]; then
    fail "Unsupported module '${module_name}'. This script must be run inside a BgScan project."
  fi

  success "Project validated — module: ${module_name}"
}

# ==============================================================================
# BUILDER INSTALLATION
# ==============================================================================
ensure_builder() {
  log "Checking for bgscan-builder binary"

  if [ -f "$BUILDER_BINARY" ]; then
    success "bgscan-builder already present at ${BUILDER_BINARY}"
    return
  fi

  warn "bgscan-builder not found at ${BUILDER_BINARY}"
  info "Looking for installer script: ${INSTALL_SCRIPT}"

  [ -f "$INSTALL_SCRIPT" ] ||
    fail "Installer script not found at ${INSTALL_SCRIPT}. Cannot proceed without it."

  info "Granting execute permission to installer script"
  chmod +x "$INSTALL_SCRIPT"

  info "Running installer: ${INSTALL_SCRIPT}"
  "$INSTALL_SCRIPT"

  [ -f "$BUILDER_BINARY" ] ||
    fail "Installer completed but ${BUILDER_BINARY} still not found"

  success "bgscan-builder installed successfully"
}

# ==============================================================================
# DEPENDENCY SETUP
# ==============================================================================
setup_dependencies() {
  log "Preparing dependency bootstrap"

  info "Granting execute permission to ${BUILDER_BINARY}"
  chmod +x "$BUILDER_BINARY"

  info "Invoking: ${BUILDER_BINARY} setup-dev --project-dir ${PROJECT_ROOT}"
  "$BUILDER_BINARY" \
    setup-dev \
    --project-dir "$PROJECT_ROOT"

  success "Dependency bootstrap finished without errors"
}

# ==============================================================================
# MAIN
# ==============================================================================
main() {
  log "Starting BgScan dependency installation"
  info "Project root: ${PROJECT_ROOT}"

  validate_project
  ensure_builder
  setup_dependencies

  local end_time elapsed
  end_time=$(date +%s)
  elapsed=$((end_time - START_TIME))

  echo
  echo -e "${C_GREEN}════════════════════════════════════════════════════════════${C_RESET}"
  echo -e "${C_GREEN}  BgScan dependency installation completed successfully${C_RESET}"
  echo -e "${C_GREEN}  Total steps : ${STEP}${C_RESET}"
  echo -e "${C_GREEN}  Elapsed time: ${elapsed}s${C_RESET}"
  echo -e "${C_GREEN}  Finished at : $(_timestamp)${C_RESET}"
  echo -e "${C_GREEN}════════════════════════════════════════════════════════════${C_RESET}"
  echo
}

main "$@"
