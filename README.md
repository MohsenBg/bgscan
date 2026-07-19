<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="./images/logo-light.svg">
  <img src="./images/logo-light.svg" alt="BGSCAN" width="520" style="max-width:100%;">
</picture>

Blazing-fast multi-protocol IP scanner with modular chain architecture

[English](./README.md) | [فارسی](./README.fa.md)

---

[![Go Version](https://img.shields.io/badge/Go-1.26.3+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-6366f1?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20|%20Windows%20|%20macOS%20|%20Termux-64748b?style=flat-square)](https://github.com/MohsenBg/bgscan/releases)
[![UI](https://img.shields.io/badge/UI-BubbleTea%20TUI-ec4899?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-22c55e?style=flat-square)](https://github.com/MohsenBg/bgscan/releases/latest)

[Installation](#installation) · [Quick Start](#quick-start) · [Documentation](#documentation) · [Protocols](#supported-protocols)

</div>

## Overview

**bgscan** is a terminal-based multi-protocol network scanner written in Go. It probes IP addresses across ICMP, TCP, HTTP, DNS, and Xray protocols and chains scan stages into pipelines — all through an interactive keyboard-driven TUI.

Use it for host discovery, port scanning, web service testing, DNS resolver analysis, tunneling detection, and Xray proxy validation. Results are saved to disk and can be re-scanned, so you can run a broad ICMP sweep and then drill down with TCP or HTTP on whatever responded.

The scanner is built for developers and researchers who need speed, flexibility, and a modern terminal experience without leaving the keyboard.

<img width="1258" height="690" alt="bgscan" src="https://github.com/user-attachments/assets/998c2c7c-f960-4a71-a022-72d86b13c6fb" />

---

## Features

**Scanning engine**

- **Multi-protocol probes** — ICMP, TCP, HTTP/1.1, HTTP/2, HTTP/3 (QUIC), TLS, DNS, DNSTT, Slipstream, and Xray
- **Pipeline chaining** — chain stages (e.g. ICMP → TCP → HTTP) with Stream, Sequential, or Batch execution modes
- **Concurrent workers** — configurable per-probe worker pools for parallel scanning
- **Shuffle & sample** — randomize target order and cap the number of IPs to scan

**Terminal UI**

- **BubbleTea TUI** — fully keyboard-driven; scan, monitor, and manage results without a browser
- **Live scan dashboard** — tabbed progress bars and result tables per pipeline stage
- **In-app settings inspector** — edit every configuration field interactively; changes save to disk immediately
- **Theme support** — dark/light/auto palettes (Catppuccin-based)

**Data & I/O**

- **Save & replay** — results are written to CSV and can be used as input for new scans
- **Bundled IP lists** — ships with Cloudflare, AWS, Azure, Google, Akamai, Fastly, Bunny, G-Core, and Iran ranges
- **Xray outbound manager** — add outbounds from share links or JSON files, then validate and speed-test them

**Reliability**

- **Crash-safe result writer** — atomic merge with `fsync` and temp-file rename; constant memory
- **Log rotation** — three log streams (core, UI, debug) with 50 MB rotation, 3 backups, 7-day retention
- **Startup health checks** — validates config, locates optional binaries, and warns instead of crashing when something is missing

---

## Why bgscan?

- **One tool, full chain.** Most scanners do one protocol well. bgscan chains them: ping a range, connect to survivors, HTTP-probe the open ports — in a single run.
- **Pipeline-agnostic engine.** The engine doesn't know what probes do. It feeds IPs, collects results, and flushes to disk. Adding a new scan type means implementing a three-method `Probe` interface and registering a stage builder.
- **No runtime clutter.** ICMP, TCP, and HTTP probes use the Go standard library. Xray, DNSTT, and Slipstream are optional external binaries validated at startup — missing ones log a warning and disable only their scan type.
- **File-first configuration.** All settings live in plain TOML. The in-app inspector reads and writes the same files, so what you see in the TUI is what's on disk.
- **Built for the terminal.** Keyboard navigation, overlay dialogs, live progress, streaming logs — no browser, no Electron, no web server.

---

## Supported Protocols

| Protocol | Layer | Description |
|:--------:|:-----:|-------------|
| **ICMP** | 3 | Host discovery and reachability via Ping |
| **TCP** | 4 | Connection scanning and TCP handshake validation |
| **HTTP** | 7 | HTTP/1.1, HTTP/2, and HTTP/3 (QUIC) via ALPN |
| **TLS** | 7 | TLS 1.0 through TLS 1.3 |
| **DNS** | 7 | Advanced DNS queries (UDP, TCP, DNS-over-TLS) with fallback and anti-hijacking checks |
| **DNSTT** | 7 | DNS Tunnel validation (SOCKS, no auth) |
| **Slipstream** | 7 | Slipstream tunnel validation (SOCKS, no auth) |
| **Xray** | 7 | Xray outbound validation and bandwidth speed testing |

---

## Installation

### Quick install

**Linux / macOS / Termux**

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**Android (Termux)**

```bash
pkg update -y && pkg install bash curl unzip -y
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

The installer detects your platform, downloads the latest release, extracts it to `bgscan/`, and makes the binary executable. On re-run it detects an existing install and offers to replace it or back it up.

### Manual install

1. Download the ZIP for your platform from the [Releases page](https://github.com/MohsenBg/bgscan/releases/latest).
2. Extract the archive.
3. Run the binary:
   - **Linux / macOS / Termux:** `./bgscan`
   - **Windows:** `bgscan.exe`

The first run creates a `settings/` directory with default TOML config and an `ips/` directory with bundled IP lists.

### Build from source

bgscan uses a companion builder tool (`bgscan-builder`) to fetch platform-specific dependencies (Xray, DNSTT, Slipstream binaries) and build the project.

```bash
git clone https://github.com/MohsenBg/bgscan.git
cd bgscan

# Install the builder tool
# Linux / macOS:
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh | bash
# Windows (PowerShell):
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1 | iex

# Fetch dependencies for your platform
# Linux / macOS:
./scripts/install-deps.sh
# Windows:
./scripts/install-deps.ps1

# Build and run
go run ./cmd/bgscan/
```

For release builds targeting a specific platform:

```bash
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os darwin -arch arm64
```

> **Note:** bgscan cannot be installed via `go install` because it requires platform-specific external binaries (Xray, DNSTT, Slipstream). Use the builder tool or the quick install script.

---

## Quick Start

1. Launch bgscan from your installation folder (`./bgscan` on Unix, `bgscan.exe` on Windows).
2. Select **Run Scan** and press `Enter`.
3. Choose a target source — **IP List** (from your imported files) or **Result List** (from a previous scan).
4. Pick a scan type — ICMP, TCP, HTTP, DNS, or Xray.
5. Press `Enter` to start. Progress and results stream live in the dashboard.
6. Open **Result Files** from the main menu to review, rename, or delete saved results.

| Key | Navigation |
|:---:|-------------|
| `↑` `↓` | Move between items |
| `Enter` | Select / start |
| `b` or `Esc` | Go back |
| `q` | Quit |

---

## Documentation

Full documentation is available at:

- **Homepage:** https://mohsenbg.github.io/bgscan
- **Full docs:** https://mohsenbg.github.io/bgscan/docs

The documentation covers:

- **Quick Start** — installation, launch, and first scan
- **Scanner** — scan types, scan sources, IP lists, result files, scan pipeline, Xray outbounds
- **Settings** — every TOML config file explained: general, writer, ICMP, TCP, HTTP, DNS, Xray, and the in-app inspector
- **Logs** — three log streams, the log viewer, and rotation policy
- **Developer** — architecture, core (engine, probes, config, results), UI (component model, layout, theming), contributing guide, and build instructions

---

## Configuration

All configuration lives in plain TOML files in the `settings/` directory next to the binary:

| File | Purpose |
|------|---------|
| `general_settings.toml` | Pipeline mode, max IPs, batch size, shuffle, status interval |
| `writer_settings.toml` | Result buffering, flush interval, channel and batch size |
| `icmp_settings.toml` | ICMP timeout, retries, workers |
| `tcp_settings.toml` | TCP port, timeout, retries, workers |
| `http_settings.toml` | HTTP/HTTPS/HTTP3 version, TLS range, accepted status codes |
| `dns_settings.toml` | DNS resolver, DNSTT, and Slipstream tuning |
| `xray_settings.toml` | Xray connectivity test type, speed test, pre-scan |

Edit files manually or use the in-app inspector — both write to the same files. The inspector saves changes immediately to disk without restarting. See the [Settings documentation](https://mohsenbg.github.io/bgscan/docs/settings/) for every field.

---

## Example Usage

**Scan a bundled IP list**

After launching, select **Run Scan → IP List → cloudflare_IPv4**, then choose a scan type (e.g. `t` for TCP). bgscan scans the list with the configured workers and writes results to `result/tcp/`.

**Chain ICMP → TCP → HTTP**

In `general_settings.toml`, set `pipeline_mode = "streaming"`. If multiple scan types are enabled, their stages chain automatically — only IPs that pass each stage's success criteria proceed to the next.

**Re-scan a previous result**

Select **Run Scan → Result List** and pick a saved result file. bgscan re-scans only the IPs in that file, useful for deeper analysis on hosts that already passed an earlier stage.

**Xray outbound validation**

From the main menu, open **Xray → Outbounds**, press `a` to add a template from a share link (`vless://`, `vmess://`, `trojan://`, `ss://`, `hysteria2://`, `wireguard://`) or a JSON file. Then run an Xray scan to test connectivity and bandwidth.

---

## Supported Platforms

| Platform | Architectures |
|----------|--------------|
| Linux | amd64, arm64, arm32, 386 |
| Windows | amd64. arm64 (10+) |
| macOS | amd64, arm64 |
| Android (Termux) | arm64, arm32, x86_64, x86 |

> **Termux:** install from F-Droid (the Play Store version is outdated).

---

## Project Structure

```
bgscan/
├── cmd/bgscan/              # Application entry point
├── internal/
│   ├── core/
│   │   ├── config/          # TOML configuration + validators
│   │   ├── scanner/         # Scanner orchestrator, engine, probes, port manager
│   │   ├── result/          # Async writer, CSV merge, registry, loader
│   │   ├── iplist/          # IP list loader, parser, registry, shuffle
│   │   ├── dns/             # DNS query helpers, DNSTT, Slipstream, SOCKS5
│   │   ├── xray/            # Xray runner, outbound/link parsing, speed test
│   │   ├── process/         # Cross-platform process lifecycle
│   │   └── fileutil/        # CSV, JSON, TOML, text, temp-file helpers
│   ├── logger/              # Leveled logging with lumberjack rotation
│   ├── startup/             # Health checks (logger, config, xray, dnstt, slipstream)
│   └── ui/                  # BubbleTea TUI (components, menus, tables, theme)
├── assets/                  # Xray, DNSTT, Slipstream binaries + outbound templates
├── ips/                     # Bundled and imported IP lists (CSV)
├── settings/                # Default TOML configuration files
├── result/                  # Scan output (CSV, per scan type)
├── logs/                    # core.log, ui.log, debug.log
├── scripts/                 # Install, build, dependency, and release scripts
├── docs/                    # Hugo Book documentation site
└── go.mod
```

---

## Contributing

Contributions are welcome — bug fixes, documentation improvements, and new features.

- **Branch from `main`** and use a descriptive branch name (`feature/`, `fix/`, `docs/`, `refactor/`).
- **Follow existing conventions** — keep code simple and readable, prefer small focused commits.
- **Commit messages** — use `feat:`, `fix:`, `docs:`, `refactor:`, `test:` prefixes.
- **Before opening a PR** — ensure the project builds, existing tests pass, and documentation is updated for new features.
- **For large changes** — open an Issue first to discuss the approach.

See the full [Contributing Guide](https://mohsenbg.github.io/bgscan/docs/developer/contributing/) in the documentation.

---

## License

[MIT](LICENSE) — Copyright (c) 2026 Mohsen Bagheri

---

## Support / Donate

If bgscan has been useful to you, consider supporting its development:

| Network | Currency | Address |
|:-------:|:--------:|---------|
| Bitcoin | `BTC` | `bc1qdwh57dm97nmx5jzdr7lrc9cxe5xh3zc59er7z9` |
| Ethereum | `ETH` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| BNB Smart Chain | `BNB / BEP20` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| TRON | `TRX / TRC20` | `TNW6pbfY8zZVZezZWyYXo7h12MycRsVJK7` |

---

<div align="center">

Built with Go · MIT License

</div>
