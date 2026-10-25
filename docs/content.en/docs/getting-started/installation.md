---
title: "Installation"
weight: 3
bookFlatSection: false
---

# Installation

bgscan runs on Linux, macOS, Windows, and Android (Termux). Pick the method that fits your operating system.

## Quick install

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

The install scripts use `bgscan-installer`. This is a small standalone binary written in Rust and kept in a separate repository ([bgscan-installer](https://github.com/MohsenBg/bgscan-installer)).

`bgscan-installer` finds the release that fits your system, checks its SHA-256 checksum against the value recorded in `checksum.txt`, and installs bgscan inside the `bgscan/` directory. If a `bgscan` binary already exists in the current directory, the installation happens there.

If an earlier installation is found, you can choose one of these options:

- **Update:** The binary is replaced and new files are added, but the `ips`, `assets`, and `settings` directories stay untouched.
- **Clean install:** The earlier installation is replaced with a fresh one.
- **Backup:** A backup of the current installation is made in a directory named `bgscan_<timestamp>`.
- **Cancel:** The installation stops.

If you have a specific version in mind and do not want the latest release, use the `--version` option:

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh -s -- --version v2.12.0
```

## Manual install

1. Download the ZIP for your operating system and device architecture from the [Releases page](https://github.com/MohsenBg/bgscan/releases/latest).
2. Extract the ZIP file.
3. Run the program:
   - **Linux/macOS/Termux:** Open a terminal, go into the program directory, and run `./bgscan`.
   - **Windows:** Double click `bgscan.exe` or run `.\bgscan.exe` in PowerShell.

On the first run, a `settings/` directory is created with the default settings files, along with an `ips/` directory holding the bundled IP lists.

## Build from source

> **Note:** bgscan cannot be installed with `go install`, because the program needs the Slipstream library for each platform. To build the program you must use the companion tool **`bgscan-builder`**. This tool is only for building the program. Installing and updating released versions is handled by `bgscan-installer`.

> Xray does not need a separate binary. It runs inside the same process through an internal library.

#### Prerequisites

- Go 1.27 or higher
- Git

#### Clone and build

```bash
# Clone the repository
git clone --recurse-submodules https://github.com/MohsenBg/bgscan.git
cd bgscan

# Install the builder tool
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1 | iex

# Fetch the Slipstream client library and bundled IP lists for your platform
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

To update bgscan, simply run the [quick install](#quick-install) script again. The installer finds the earlier installation and gives you the options to update, do a clean install, take a backup, or cancel.

In update mode the `settings`, `ips`, and `assets` directories are kept.

If you choose a clean install, it is better to back up what you need from the earlier version first:

- Move the `settings/*.toml` files you changed to the new installation.
- Take your custom IP lists from the `ips/` directory.
- Close bgscan before replacing the program files.

## System requirements

- **Operating system:** Linux, macOS, Windows 10 or higher, or Android 7.0 or higher with Termux
- **Tools:** `curl`. On most systems the installer installs the tools it needs itself.
- **Windows:** PowerShell 5.1 or higher
- **Termux:** Install the F-Droid version. The Play Store version is old.
