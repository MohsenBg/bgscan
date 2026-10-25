---
title: "Getting Started"
weight: 1
bookFlatSection: true
bookCollapseSection: true
---

# Getting Started (Developer)

This guide walks through setting up a development environment, building **bgscan**, and running the project locally.

> **Before you start:** read the [Contributing](./contributing/) guide so you know the branch rules, how commits are written, and how pull requests work.

---

## Prerequisites

- [Go](https://go.dev/) 1.27 or newer. The exact version is set in `go.mod`.
- Git
- For Android builds: [Android NDK](https://developer.android.com/ndk)

---

## 1. Get the repository

```bash
git clone --recurse-submodules https://github.com/MohsenBg/bgscan.git
cd bgscan
```

Using `--recurse-submodules` is required. The Xray core and other upstream code live in the submodules inside `third_party/`.

## 2. Create a branch

```bash
git checkout -b feature/my-change
```

See the [Contributing](./contributing/) guide for branch naming rules.

## 3. Install dependencies

The project uses a companion tool called **`bgscan-builder`**. This tool prepares the dependencies it needs, including the Slipstream client library and the IP lists for each platform.

The install script downloads `bgscan-builder` and places it in the project root. It then uses that tool to fetch the right library for your operating system and architecture.

Xray does not need a separate binary. `xray-core` runs inside the same process.

**Linux / macOS**

```bash
./scripts/install-deps.sh
```

**Windows**

```powershell
./scripts/install-deps.ps1
```

The script does two things:

1. It downloads `bgscan-builder` into the project root.
2. It runs `bgscan-builder setup-dev --project-dir <project-root>` so the Slipstream client library for your operating system and architecture is downloaded and placed in the right directory.

## 4. Build and run the project

Once the dependencies are installed, run these commands:

```bash
go mod tidy
go run ./cmd/bgscan/
```

The [startup checks](./core/#startup) run first. If they finish without problems, press Enter to enter the TUI.

---

## Building a release

Building release output also needs `bgscan-builder`. If you did not install it during the dependency step, use the script for your operating system.

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

Android builds need the Android NDK installed. Point to it with the `-ndk-dir` flag:

```bash
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

---

## `bgscan-builder` reference

`bgscan-builder` is a small command line tool written in Go for building bgscan, and it has two subcommands:

- **`setup-dev`**: prepares the dependencies for the current operating system and architecture so you can run the project with `go run ./cmd/bgscan/`.
- **`release`**: prepares the Slipstream client library and builds the release binary for the operating system and architecture you specify.

Installing and updating released versions is not this tool's job. The separate [bgscan-installer](https://github.com/MohsenBg/bgscan-installer) tool handles that. See the [Installation](../getting-started/installation/) guide for details.

**Flags**

| Flag | Subcommand | Description |
|---|---|---|
| `-arch string` | `release` | Target architecture: `amd64`, `arm64`, `arm32`, `amd32`, or `all` |
| `-dest string` | `release` | Release output directory. The default is `./dist` |
| `-ndk-dir string` | `release` (Android) | Root path of the Android NDK |
| `-os string` | `release` | Target operating system: `linux`, `windows`, `macos`, `android`, or `all` |
| `-project-dir string` | `setup-dev`, `release` | Path to the bgscan project |
| `-verbose` | all | Show details of each step as it runs |
| `-version string` | `release` | Version recorded inside the built binary |
