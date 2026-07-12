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
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**Android (Termux)**

```bash
pkg update -y && pkg install bash curl unzip -y && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

The installer downloads the latest release, extracts it to `bgscan/`, and makes the binary executable. If a previous installation exists, it prompts you to remove it or back it up as `bgscan_old`.

## Manual Install

1. Download the ZIP for your platform from the [Releases page](https://github.com/MohsenBg/bgscan/releases/latest).
2. Extract the archive.
3. **Run the application:**
   - **Linux/macOS/Termux:** Open terminal, navigate to the folder, and run `./bgscan`.
   - **Windows:** Simply double-click `bgscan.exe` to launch, or run `.\bgscan.exe` in PowerShell.

The first run creates the default `settings/` folder with configuration files and an `ips/` folder with bundled IP lists.

## Build from Source

> **Note:** bgscan cannot be installed via `go install`. You must build it using the `bg-builder` tool included in the repository.

#### Prerequisites

- Go 1.26.3+
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

# Build for your current system
./bg-builder

# Or build for a specific target
./bg-builder --os linux --arch amd64
./bg-builder --os windows --arch amd64
./bg-builder --os darwin --arch arm64
./bg-builder --os android --arch arm64
```

## Upgrading

To upgrade, simply re-run the [Quick Install](#quick-install) script. It will detect the existing version and offer to replace it or back it up.

If you have customized configurations:

- Copy your custom `settings/*.toml` files to the new installation.
- Move any custom IP lists from `ips/`.
- Stop any running bgscan instance before replacing files.

## Requirements

- **OS:** Linux, macOS, Windows 10+, or Android 7.0+ (Termux)
- **Tools:** `curl`, `unzip`, and `bash` (installer handles missing dependencies on most systems)
- **Windows:** PowerShell 5.1+
- **Termux:** Install from F-Droid (Play Store version is outdated)
