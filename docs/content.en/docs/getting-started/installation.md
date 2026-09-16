---

title: "Installation"
weight: 3
bookFlatSection: false
---

# Installation

bgscan runs on Linux, macOS, Windows, and Android (Termux). Pick the method that fits your environment.

## Quick Install

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**Android (Termux)**

```bash
{ command -v curl >/dev/null 2>&1 || pkg install -y curl; } && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

The install scripts use `bgscan-installer`, a small standalone Rust binary from the separate [bgscan-installer](https://github.com/MohsenBg/bgscan-installer) project. It resolves the requested release asset for your platform, verifies its SHA-256 checksum against `checksum.txt`, and installs bgscan into `bgscan/` (or into the current directory if a `bgscan` binary is already there). If an installation already exists, you choose how to proceed: update in place (replaces the binary and adds missing files, keeps your `ips`, `assets`, and `settings`), clean install, back up the existing installation to a timestamped `bgscan_<timestamp>` directory, or cancel.

To install a specific version instead of the latest, pass `--version`:

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh -s -- --version v2.11.0
```

## Manual Install

1. Download the ZIP for your platform from the [Releases page](https://github.com/MohsenBg/bgscan/releases/latest).
2. Extract the archive.
3. **Run the application:**
   - **Linux/macOS/Termux:** Open terminal, navigate to the folder, and run `./bgscan`.
   - **Windows:** Simply double-click `bgscan.exe` to launch, or run `.\bgscan.exe` in PowerShell.

The first run creates the default `settings/` folder with configuration files and an `ips/` folder with bundled IP lists.

## Build from Source

> **Note:** bgscan cannot be installed via `go install` because it needs the platform-specific Slipstream sidecar binary. You must build it using the companion **`bgscan-builder`** tool. `bgscan-builder` only builds; installing and updating releases is handled by the separate `bgscan-installer` tool described above. (Xray needs no binary: it runs in-process through a vendored library.)

#### Prerequisites

- Go 1.27+
- Git

#### Clone and Build

```bash
# Clone the repository
git clone https://github.com/MohsenBg/bgscan.git
cd bgscan

# Install the builder tool
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1 | iex

# Fetch the Slipstream sidecar and bundled IP lists for your platform
# Linux/macOS
./scripts/install-deps.sh
# Windows (PowerShell)
./scripts/install-deps.ps1

# Run in development
go run ./cmd/bgscan/

# Or build release artifacts targeting a specific platform (see `bgscan-builder release -h` for all flags)
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os macos -arch arm64
./bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

## Upgrading

To upgrade, simply re-run the [Quick Install](#quick-install) script. It will detect the existing installation and offer to update, clean install, back up, or cancel.

The update option keeps your `settings`, `ips`, and `assets`. If you choose a clean install instead, back up first:

- Copy your custom `settings/*.toml` files to the new installation.
- Move any custom IP lists from `ips/`.
- Stop any running bgscan instance before replacing files.

## Requirements

- **OS:** Linux, macOS, Windows 10+, or Android 7.0+ (Termux)
- **Tools:** `curl` (installer handles missing dependencies on most systems)
- **Windows:** PowerShell 5.1+
- **Termux:** Install from F-Droid (Play Store version is outdated)
