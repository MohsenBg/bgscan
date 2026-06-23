#!/usr/bin/env bash
# ==============================================================================
# bgscan release generator (CI/CD ONLY)
# ------------------------------------------------------------------------------
# ⚠️ DO NOT RUN MANUALLY
#
# This script is designed for GitHub Actions only.
#
# PURPOSE:
#   - Reads ZIP assets from ./dist
#   - Generates SHA256 checksum manifest
#   - Builds GitHub Release markdown (release_notes.md)
#   - Normalizes OS / ARCH naming into readable tables
#
# OUTPUT:
#   release/
#     ├── checksum.txt
#     └── release_notes.md
# ==============================================================================

set -euo pipefail

# ==============================================================================
# INPUT
# ==============================================================================
if [[ -z "${1:-}" ]]; then
  echo "Usage: $0 <tag_version>"
  exit 1
fi

TAG_VERSION="$1"

ROOT_DIR="$PWD"
DIST_DIR="$ROOT_DIR/dist"
RELEASE_DIR="$ROOT_DIR/release"

CHECKSUM_FILE="$RELEASE_DIR/checksum.txt"
NOTES_FILE="$RELEASE_DIR/release_notes.md"

REPO_URL="${GITHUB_SERVER_URL:-https://github.com}/${GITHUB_REPOSITORY:-user/repo}"

mkdir -p "$RELEASE_DIR"

# ==============================================================================
# LOGGER
# ==============================================================================
log() {
  echo
  echo "============================================================"
  echo "$*"
  echo "============================================================"
}

# ==============================================================================
# MAPS (SOURCE OF TRUTH)
# ==============================================================================
declare -A OS_MAP=(
  ["linux"]="🐧 Linux"
  ["windows"]="🪟 Windows"
  ["android"]="🤖 Android"
  ["macos"]="🍏 macOS"
)

declare -A ARCH_MAP=(
  # Linux
  ["linux-64"]="AMD64 / x64"
  ["linux-32"]="x86 / 32-bit"
  ["linux-arm64"]="ARM64"
  ["linux-arm32-v7a"]="ARM32 (ARMv7)"

  # Windows
  ["windows-64"]="AMD64 / x64"
  ["windows-arm64"]="ARM64"

  # Android
  ["android-arm64-v8a"]="ARM64 (v8a)"
  ["android-armeabi-v7a"]="ARM32 (armeabi-v7a)"
  ["android-x86"]="x86 / 32-bit"
  ["android-x86_64"]="AMD64 / x64"

  # macOS
  ["macos-64"]="AMD64 / Intel x64"
  ["macos-arm64"]="ARM64 / Apple Silicon"
)

# ==============================================================================
# VALIDATION
# ==============================================================================
log "Checking dist directory"

if [[ ! -d "$DIST_DIR" ]] || [[ -z "$(ls -A "$DIST_DIR" 2>/dev/null)" ]]; then
  echo "ERROR: dist/ is empty"
  exit 1
fi

# ==============================================================================
# CHECKSUM GENERATION
# ==============================================================================
log "Generating SHA256 checksums"

: >"$CHECKSUM_FILE"
cd "$DIST_DIR"

FILES=()

for file in *; do
  [[ -f "$file" ]] || continue
  sha256sum "$file" >>"$CHECKSUM_FILE"
  FILES+=("$file")
done

# ==============================================================================
# RELEASE NOTES START
# ==============================================================================
log "Generating release notes"

: >"$NOTES_FILE"

cat <<EOF >>"$NOTES_FILE"
# 🚀 bgscan Release $TAG_VERSION

Automated multi-platform release generated via GitHub Actions.

All artifacts are **prebuilt ZIP packages**.

---

## 📦 Download Table

| 🌍 Platform | 🧬 Architecture | 📥 Download |
|------------|----------------|------------|
EOF

# ==============================================================================
# TABLE GENERATION (SMART PARSER)
# ==============================================================================
for file in "${FILES[@]}"; do

  # remove extension
  clean="${file%.zip}"
  clean="${clean%.exe}"

  # remove prefix
  key="${clean#bgscan-}"

  OS_KEY="${key%%-*}"
  ARCH_KEY="${key#*-}"

  FULL_KEY="${OS_KEY}-${ARCH_KEY}"

  OS_NAME="${OS_MAP[$OS_KEY]:-❓ Unknown}"
  ARCH_NAME="${ARCH_MAP[$FULL_KEY]:-$ARCH_KEY}"

  LINK="[$file]($REPO_URL/releases/download/$TAG_VERSION/$file)"

  echo "| $OS_NAME | $ARCH_NAME | $LINK |" >>"$NOTES_FILE"

done

# ==============================================================================
# CHECKSUM TABLE (CLEAN + PROFESSIONAL)
# ==============================================================================
cat <<EOF >>"$NOTES_FILE"

---

## 🔐 SHA256 Checksums

| Asset | SHA256 |
|-------|--------|
EOF

while read -r hash file; do
  printf "| %s | \`%s\` |\n" "$file" "$hash" >>"$NOTES_FILE"
done <"$CHECKSUM_FILE"

# ==============================================================================
# FOOTER
# ==============================================================================
cat <<EOF >>"$NOTES_FILE"

---

## ⚙️ Build Info

- Project: bgscan
- Version: $TAG_VERSION
- Pipeline: GitHub Actions CI/CD
EOF

# ==============================================================================
# DONE
# ==============================================================================
log "Release generation completed"
echo "Output:"
ls -lh "$RELEASE_DIR"
