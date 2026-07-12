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

- [Go](https://go.dev/) 1.26.3+ (see `go.mod` for the exact version)
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

bgscan uses a companion tool called **`bgscan-builder`** to fetch and build the project's dependencies. The install scripts download it for you, place it in the project root, and use it to fetch the correct dependency build for your OS/architecture.

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
2. Run `bgscan-builder setup-dev --project-dir <project-root>` to download the correct dependencies for your OS/arch and place them in the right directory.

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

Builds release artifacts for one or more OS/architecture combinations.

**Flags:**

| Flag | Description |
|---|---|
| `-arch string` | Target architecture (`amd64`, `arm64`, `arm32`, `amd32`, `all`) |
| `-dep-version string` | Dependencies version tag (default `"v1.0"`) |
| `-dest string` | Release output directory (default `"./dist"`) |
| `-ndk-dir string` | Android NDK root directory |
| `-os string` | Target operating system (`linux`, `windows`, `macos`, `android`, `all`) |
| `-project-dir string` | Path to the bgscan project |
| `-xray-version string` | Xray version tag (default `"v26.3.27"`) |
