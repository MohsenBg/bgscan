---
title: "Getting Started"
weight: 1
bookFlatSection: true
bookCollapseSection: true
---

# Getting Started (Developer)

This guide explains how to set up your environment, build, and run **bgscan** locally for development.

> **Read first:** [Contributing](../contributing/) — branching, commit conventions, and PR workflow.

---

## Prerequisites

- [Go](https://go.dev/) 1.27+ (see `go.mod` for the exact version)
- Git
- For Android builds: [Android NDK](https://developer.android.com/ndk)

---

## 1. Clone the repository

```bash
git clone https://github.com/MohsenBg/bgscan.git
cd bgscan
```

## 2. Create a branch

```bash
git checkout -b feature/my-change
```

See [Contributing](../contributing/) for branch naming conventions.

## 3. Install dependencies

bgscan uses a companion tool called **`bgscan-builder`** to fetch the platform-specific Slipstream sidecar binary and bundled IP lists, and to build release artifacts. The install scripts download it for you, place it in the project root, and use it to fetch the correct sidecar build for your OS/architecture. (Xray needs no binary: it runs in-process through a vendored xray-core library.)

**Linux / macOS**

```bash
./scripts/install-deps.sh
```

**Windows**

```powershell
./scripts/install-deps.ps1
```

This script will:

1. Download `bgscan-builder` into the project root.
2. Run `bgscan-builder setup-dev --project-dir <project-root>` to download the Slipstream sidecar for your OS/arch and place it in the right directory.

## 4. Build and run

Once dependencies are installed:

```bash
go mod tidy
go run ./cmd/bgscan/
```

The [startup health checks](../core/#startup) run first. Once they pass, press Enter to enter the TUI.

---

## Building Releases

To build release artifacts, you also need `bgscan-builder`. If you don't already have it from the dependency step, install it:

**Linux / macOS**

```bash
./scripts/install-builder.sh
```

**Windows**

```powershell
./scripts/install-builder.ps1
```

#### Build commands

```bash
bgscan-builder release -os linux -arch amd64
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
bgscan-builder release -os all -arch all -dest ./dist
```

#### Building for Android

Android builds require the Android NDK. Pass its path with `-ndk-dir`:

```bash
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

---

## bgscan-builder reference

A small Go CLI for building the project. It has two subcommands:

- **setup-dev** — download the correct dependency binaries for the host so `go run ./cmd/bgscan/` works locally.
- **release** — stage the Slipstream sidecar and build a release binary for a target OS/arch.

Installing and updating released builds is not part of the builder. That is handled by the separate [bgscan-installer](https://github.com/MohsenBg/bgscan-installer) tool (see [Installation](../../getting-started/installation/)).

**Flags:**

| Flag | Subcommand | Description |
|---|---|---|
| `-arch string` | release | Target architecture (`amd64`, `arm64`, `arm32`, `amd32`, `all`) |
| `-dest string` | release | Release output directory (default `"./dist"`) |
| `-ndk-dir string` | release (android) | Android NDK root directory |
| `-os string` | release | Target operating system (`linux`, `windows`, `macos`, `android`, `all`) |
| `-project-dir string` | setup-dev, release | Path to the bgscan project |
| `-verbose` | all | Print each step as it runs |
| `-version string` | release | Version embedded into the built binary |
