#!/usr/bin/env bash
# ============================================================
#  bgscan installer
#  https://github.com/MohsenBg/bgscan
# ============================================================
set -euo pipefail

# ── colour palette ──────────────────────────────────────────
if [ -t 1 ]; then
  BOLD="\033[1m"
  DIM="\033[2m"
  RESET="\033[0m"
  RED="\033[38;5;196m"
  GREEN="\033[38;5;82m"
  CYAN="\033[38;5;51m"
  YELLOW="\033[38;5;220m"
  BLUE="\033[38;5;39m"
  MAGENTA="\033[38;5;171m"
  WHITE="\033[38;5;255m"
  GREY="\033[38;5;245m"
else
  BOLD="" DIM="" RESET="" RED="" GREEN="" CYAN=""
  YELLOW="" BLUE="" MAGENTA="" WHITE="" GREY=""
fi

REPO="MohsenBg/bgscan"
API="https://api.github.com/repos/$REPO/releases/latest"

# ── helpers ─────────────────────────────────────────────────
print_banner() {
  echo -e ""
  echo -e "${CYAN}${BOLD}  ██████╗  ██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗${RESET}"
  echo -e "${CYAN}${BOLD}  ██╔══██╗██╔════╝ ██╔════╝██╔════╝██╔══██╗████╗  ██║${RESET}"
  echo -e "${CYAN}${BOLD}  ██████╔╝██║  ███╗███████╗██║     ███████║██╔██╗ ██║${RESET}"
  echo -e "${CYAN}${BOLD}  ██╔══██╗██║   ██║╚════██║██║     ██╔══██║██║╚██╗██║${RESET}"
  echo -e "${CYAN}${BOLD}  ██████╔╝╚██████╔╝███████║╚██████╗██║  ██║██║ ╚████║${RESET}"
  echo -e "${CYAN}${BOLD}  ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝${RESET}"
  echo -e ""
  echo -e "  ${GREY}Installer  •  github.com/${REPO}${RESET}"
  echo -e "  ${GREY}────────────────────────────────────────────────────${RESET}"
  echo -e ""
}

step()    { echo -e "  ${BLUE}${BOLD}→${RESET}  ${WHITE}$*${RESET}"; }
ok()      { echo -e "  ${GREEN}${BOLD}✔${RESET}  ${GREEN}$*${RESET}"; }
warn()    { echo -e "  ${YELLOW}${BOLD}⚠${RESET}  ${YELLOW}$*${RESET}"; }
info()    { echo -e "  ${GREY}   $*${RESET}"; }
fatal()   {
  echo -e ""
  echo -e "  ${RED}${BOLD}✖  Error:${RESET} ${RED}$*${RESET}"
  echo -e ""
  exit 1
}

section() {
  echo -e ""
  echo -e "  ${MAGENTA}${BOLD}$*${RESET}"
  echo -e "  ${GREY}$(printf '─%.0s' {1..48})${RESET}"
}

prompt_choice() {
  local _question="$1"; shift
  local _options=("$@")
  echo -e ""
  echo -e "  ${WHITE}${BOLD}$_question${RESET}"
  echo -e ""
  local i=1
  for opt in "${_options[@]}"; do
    echo -e "    ${CYAN}${BOLD}[$i]${RESET}  ${WHITE}$opt${RESET}"
    (( i++ ))
  done
  echo -e ""
  printf "  ${BOLD}Your choice:${RESET} "
  read -r _choice
  echo "$_choice"
}

spinner() {
  local pid=$1
  local label="${2:-Working}"
  local delay=0.08
  local frames=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
  local i=0
  tput civis 2>/dev/null || true
  while kill -0 "$pid" 2>/dev/null; do
    printf "\r  ${CYAN}${BOLD}%s${RESET}  ${WHITE}%s${RESET}  " "${frames[$i]}" "$label"
    i=$(( (i + 1) % ${#frames[@]} ))
    sleep "$delay"
  done
  printf "\r%-60s\r" " "
  tput cnorm 2>/dev/null || true
}

# ── detect platform ─────────────────────────────────────────
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux*)  OS="linux"  ;;
    Darwin*) OS="macos"  ;;
    *)       OS="unknown";;
  esac

  # Termux check
  if [ -n "${PREFIX:-}" ] && echo "${PREFIX:-}" | grep -q "com.termux"; then
    OS="termux"
  fi
}

# ── dependency: unzip ───────────────────────────────────────
ensure_unzip() {
  if command -v unzip >/dev/null 2>&1; then
    ok "unzip is available"
    return
  fi

  warn "unzip not found — attempting to install..."

  if   command -v apt   >/dev/null 2>&1; then sudo apt-get update -qq && sudo apt-get install -y unzip
  elif command -v apk   >/dev/null 2>&1; then sudo apk add --no-cache unzip
  elif command -v yum   >/dev/null 2>&1; then sudo yum install -y unzip
  elif command -v dnf   >/dev/null 2>&1; then sudo dnf install -y unzip
  elif command -v pkg   >/dev/null 2>&1; then pkg install unzip -y
  elif [ "$OS" = "macos" ]; then
    fatal "unzip missing. Install Xcode Command Line Tools:\n\n    xcode-select --install"
  else
    fatal "Cannot install unzip automatically. Please install it and re-run."
  fi

  ok "unzip installed"
}

# ── resolve asset name ───────────────────────────────────────
resolve_asset() {
  case "$OS" in
    linux)
      case "$ARCH" in
        x86_64)           echo "bgscan-linux-64.zip"         ;;
        aarch64|arm64)    echo "bgscan-linux-arm64.zip"      ;;
        armv7l|armv7*)    echo "bgscan-linux-arm32-v7a.zip"  ;;
        i386|i686)        echo "bgscan-linux-32.zip"         ;;
        *) return 1 ;;
      esac ;;
    macos)
      case "$ARCH" in
        arm64)   echo "bgscan-macos-arm64.zip" ;;
        x86_64)  echo "bgscan-macos-64.zip"    ;;
        *) return 1 ;;
      esac ;;
    termux)
      case "$ARCH" in
        aarch64|arm64)  echo "bgscan-android-arm64-v8a.zip"  ;;
        armv7l|armv7*)  echo "bgscan-android-armeabi-v7a.zip";;
        x86_64)         echo "bgscan-android-x86_64.zip"     ;;
        i686|x86)       echo "bgscan-android-x86.zip"        ;;
        *) return 1 ;;
      esac ;;
    *) return 1 ;;
  esac
}

# ── fetch latest release download URL ───────────────────────
fetch_download_url() {
  local asset="$1"
  curl -fsSL "$API" \
    | grep -o '"browser_download_url": *"[^"]*"' \
    | grep "$asset" \
    | cut -d '"' -f4
}

# ── progress download with curl ─────────────────────────────
download_with_progress() {
  local url="$1"
  local dest="$2"
  curl -L --progress-bar "$url" -o "$dest" 2>&1 \
    | while IFS= read -r line; do
        printf "  ${GREY}%s${RESET}\n" "$line"
      done
  # also works without progress pipe — just use curl directly if terminal
  curl -L --progress-bar "$url" -o "$dest"
}

# ════════════════════════════════════════════════════════════
#  MAIN
# ════════════════════════════════════════════════════════════
print_banner

# ── 1. Detect platform ──────────────────────────────────────
section "Detecting environment"
detect_platform

info "Operating system : ${BOLD}$OS${RESET}"
info "Architecture     : ${BOLD}$ARCH${RESET}"

ASSET="$(resolve_asset)" || fatal "Unsupported platform: $OS / $ARCH\n\n  Please open an issue: https://github.com/$REPO/issues"
info "Release asset    : ${BOLD}$ASSET${RESET}"
ok "Platform supported"

# ── 2. Resolve install location ─────────────────────────────
section "Install location"
case "$OS" in
  termux) INSTALL_BASE="/data/data/com.termux/files/usr/bin" ;;
  *)      INSTALL_BASE="/usr/local/bin" ;;
esac
INSTALL_DIR="$INSTALL_BASE/bgscan"
info "Target : ${BOLD}$INSTALL_DIR${RESET}"

# ── 3. Handle existing install ───────────────────────────────
if [ -d "$INSTALL_DIR" ]; then
  echo ""
  warn "An existing installation was found at ${BOLD}$INSTALL_DIR${RESET}"

  CHOICE="$(prompt_choice \
    "How would you like to proceed?" \
    "Remove the old installation and install fresh" \
    "Back up the old installation (rename to bgscan_old)" \
    "Cancel — exit the installer")"

  case "$CHOICE" in
    1)
      step "Removing old installation..."
      rm -rf "$INSTALL_DIR"
      ok "Old installation removed"
      ;;
    2)
      BACKUP="${INSTALL_DIR}_old"
      step "Backing up to ${BACKUP}..."
      [ -d "$BACKUP" ] && rm -rf "$BACKUP"
      mv "$INSTALL_DIR" "$BACKUP"
      ok "Backup saved to $BACKUP"
      ;;
    *)
      echo ""
      info "Installation cancelled. No changes were made."
      echo ""
      exit 0
      ;;
  esac
fi

mkdir -p "$INSTALL_DIR"

# ── 4. Ensure unzip is available ────────────────────────────
section "Checking dependencies"
ensure_unzip

# ── 5. Fetch download URL ───────────────────────────────────
section "Fetching release information"
step "Querying GitHub API..."

URL="$(fetch_download_url "$ASSET")"
[ -z "$URL" ] && fatal "Could not find a download URL for \"$ASSET\".\n\n  Check releases: https://github.com/$REPO/releases"

ok "Release URL resolved"
info "$URL"

# ── 6. Download ─────────────────────────────────────────────
section "Downloading"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

step "Downloading $ASSET..."
echo ""
(curl -L --progress-bar "$URL" -o "$TMP/bgscan.zip") 2>&1 | \
  sed "s/^/  /"
echo ""
ok "Download complete"

# ── 7. Extract & install ─────────────────────────────────────
section "Installing"
step "Extracting archive..."
unzip -q "$TMP/bgscan.zip" -d "$TMP/extracted"

BIN="$(find "$TMP/extracted" -type f -name "bgscan*" | head -n 1)"
[ -z "$BIN" ] && fatal "Binary not found inside the archive. The release may be corrupted."

chmod +x "$BIN"
mv "$BIN" "$INSTALL_DIR/bgscan"
ok "Binary installed at ${BOLD}$INSTALL_DIR/bgscan${RESET}"

# ── 8. Done ──────────────────────────────────────────────────
echo ""
echo -e "  ${GREY}────────────────────────────────────────────────────${RESET}"
echo -e ""
echo -e "  ${GREEN}${BOLD}✔  bgscan installed successfully!${RESET}"
echo -e ""
echo -e "  ${WHITE}Get started:${RESET}"
echo -e "  ${CYAN}cd  bgscan  ${RESET}"
echo -e "  ${CYAN}./bgscan ${RESET} "
echo -e ""
echo -e "  ${GREY}Docs & source: https://github.com/${REPO}${RESET}"
echo -e "  ${GREY}────────────────────────────────────────────────────${RESET}"
echo -e ""