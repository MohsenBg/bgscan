<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="./images/logo-light.svg">
  <img src="./images/logo-light.svg" alt="BGSCAN" width="520" style="max-width:100%;">
</picture>

multi-protocol IP scanner with chainable scan stages

[English](./README.md) | [فارسی](./README.fa.md)

---

[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-GPL--3.0-6366f1?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20|%20Windows%20|%20macOS%20|%20Termux-64748b?style=flat-square)](https://github.com/MohsenBg/bgscan/releases)
[![UI](https://img.shields.io/badge/UI-BubbleTea%20TUI-ec4899?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-22c55e?style=flat-square)](https://github.com/MohsenBg/bgscan/releases/latest)

[Installation](#installation) · [Quick Start](#quick-start) · [Documentation](#documentation) · [Protocols](#supported-protocols)

</div>

## Overview

**bgscan** is a network scanner for the terminal, written in Go. It probes IPs over ICMP, TCP, HTTP, DNS, Xray, and DNS tunnels (DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, TheFeed). Those probes run on their own or as chained stages, all from a keyboard-driven TUI.

Chaining is the part worth trying first. Run a broad ICMP sweep, then send only the hosts that answered into TCP or HTTP. Results go to disk in a form the next scan can read, so each pass narrows the list.

<div align="center">
  <video
    src="https://github.com/user-attachments/assets/4a4c216e-0179-427c-8fe1-89a8fa197730"
    autoplay
    muted
    loop
    playsinline
    controls
    width="100%"
    style="max-width: 900px; border-radius: 12px; border: 1px solid #30363d;"
  ></video>
</div>

---

## Features

**Scanning engine**

- Multi-protocol probes: ICMP, TCP, HTTP/1.1, HTTP/2, HTTP/3 (QUIC), DNS, DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, TheFeed, and Xray
- Pipeline chaining between stages (for example ICMP → TCP → HTTP) in Streaming or Batch mode
- Per-probe worker pools for concurrent scanning
- Shuffle and sample: randomize target order and cap the number of IPs to scan
- Dual-stack aware: IPv4 and IPv6 targets and lists are supported end to end

**Terminal UI**

- A BubbleTea TUI driven entirely by the keyboard, down to results management
- Live scan dashboard with tabbed progress bars and per-stage result tables
- An in-app inspector for every config field, saved to disk as soon as you change it
- Dark, light, and auto themes built on Catppuccin

**Data and I/O**

- Results go to CSV files you can feed back in as a scan source
- Bundled IP lists for Cloudflare (IPv4 and IPv6), AWS, Azure, Google, Akamai, Fastly, Bunny, G-Core, Iran, and public DNS servers
- Xray outbound manager: add outbounds from share links or JSON files, then validate and speed-test them

**Reliability**

- The result writer is crash-safe: atomic merge, `fsync`, temp-file rename, constant memory
- Log rotation across three streams (core, UI, debug) with 50 MB rotation, 3 backups, and 7-day retention
- Startup health checks validate the config and look for optional binaries. If something is missing, you get a warning instead of a crash

---

## Why bgscan?

Most scanners are good at one protocol. bgscan chains them: ping a range, connect to whatever answered, HTTP-probe the ports that opened, all in one run.

The engine doesn't care what the stages do. It moves IPs through, collects results, and writes them to disk. Adding a scan type means implementing a four-method `Probe` interface (Init, Run, Schema, Close) and registering a stage builder.

There isn't much outside the binary. ICMP, TCP, HTTP, and DNS use the Go standard library. Slipstream is the only external binary, and it's checked at startup: when it's missing you get a warning and only that scan type is disabled. Xray and the DNS tunnels (DNSTT, VayDNS, MasterDNS, StormDNS, TheFeed) run in-process from vendored libraries.

Settings are plain TOML. The in-app inspector reads and writes the same files, so the TUI shows what's actually on disk.

Everything happens in the terminal: keyboard navigation, overlay dialogs, live progress, streaming logs. There's no browser and no web server behind it.

---

## Supported Protocols

| Protocol | Layer | Description |
| :---: | :---: | :--- |
| ICMP | 3 | Host discovery and reachability via Ping (IPv4 and IPv6) |
| TCP | 4 | Connection scanning and TCP handshake validation |
| HTTP | 7 | HTTP/1.1, HTTP/2, and HTTP/3 over QUIC via ALPN |
| DNS | 7 | Advanced DNS queries (UDP, TCP, DNS-over-TLS) with fallback and anti-hijacking checks |
| DNSTT | 7 | DNS tunnel validation with SOCKS and SSH base connections |
| VayDNS | 7 | DNS tunnel validation with SOCKS and SSH base connections |
| Slipstream | 7 | DNS tunnel validation with SOCKS and SSH base connections |
| MasterDNS | 7 | DNS tunnel validation with configurable encryption and MTU discovery |
| StormDNS | 7 | DNS tunnel validation with configurable encryption and MTU discovery |
| TheFeed | 7 | Encrypted DNS feed tunnel validation over UDP, TCP, or DNS-over-TLS |
| Xray | 7 | Xray outbound validation and bandwidth speed testing |

---

## Installation

### Quick install

Linux / macOS :

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

Android (Termux):

```bash
{ command -v curl >/dev/null 2>&1 || pkg install -y curl; } && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

The installer downloads the `bgscan-installer` tool (a small standalone Rust binary from the separate [bgscan-installer](https://github.com/MohsenBg/bgscan-installer) project). That tool picks the latest release asset for your platform, verifies its SHA-256 checksum, and installs bgscan into `bgscan/`. Run it again later and it detects the existing install, then offers three choices: update in place (keeps your `ips`, `assets`, and `settings`, replaces the binary, adds missing files), do a clean install, or back up the old installation to a timestamped `bgscan_<timestamp>` directory.

To install a specific version instead of the latest:

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh -s -- --version v2.11.0
```

### Manual install

1. Download the ZIP for your platform from the [Releases page](https://github.com/MohsenBg/bgscan/releases/latest).
2. Extract the archive.
3. Run the binary:
   - Linux / macOS / Termux: `./bgscan`
   - Windows: `bgscan.exe`

The first run creates a `settings/` directory with default TOML config and an `ips/` directory with bundled IP lists.

### Build from source

bgscan uses a companion build tool, `bgscan-builder`, to fetch the platform-specific Slipstream sidecar binary and bundled IP lists, and to build the project. It only handles builds; installing and updating releases is the job of the separate [`bgscan-installer`](https://github.com/MohsenBg/bgscan-installer) tool described above. Xray needs no binary: it runs in-process through a vendored library.

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

For release builds targeting a specific platform (see `bgscan-builder release -h` for all flags):

```bash
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os macos -arch arm64
./bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

> bgscan cannot be installed via `go install` because it needs the platform-specific Slipstream sidecar binary. Use the builder tool or the quick install script.

---

## Quick Start

1. Launch bgscan from your installation folder (`./bgscan` on Unix, `bgscan.exe` on Windows).
2. Select **Run Scan** and press `Enter`.
3. Choose a target source: **IP List** (from your imported files) or **Result List** (from a previous scan).
4. Pick a scan type: ICMP, TCP, HTTP, DNS Resolve, DNS Tunneling, or Xray.
5. Press `Enter` to start. Progress and results stream live in the dashboard.
6. Open **Result Files** from the main menu to review, rename, or delete saved results.

| Key | Navigation |
| :---: | :--- |
| `↑` `↓` | Move between items |
| `Enter` | Select / start |
| `b` or `Esc` | Go back |
| `q` | Quit |

---

## Documentation

Full documentation lives at:

- Homepage: <https://mohsenbg.github.io/bgscan>
- Full docs: <https://mohsenbg.github.io/bgscan/docs>

The documentation covers:

- Quick Start: installation, launch, and your first scan
- Scanner: scan types and sources, IP lists, result files, pipelines, Xray outbounds
- DNS Tunneling: setup for each protocol (DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, TheFeed), scan flow, and platform defaults
- Settings: every TOML file explained (general, writer, ICMP, TCP, HTTP, DNS, Xray) and the in-app inspector
- Logs: the three streams, the log viewer, and the rotation policy
- Developer: architecture, core packages (engine, probes, config, results), UI (components, layout, themes), contributing guide, and build instructions

---

## Configuration

All configuration lives in plain TOML files in the `settings/` directory next to the binary:

| File | Purpose |
| :--- | :--- |
| `general_settings.toml` | Pipeline mode, batch size, max IPs per stage, max successful IPs, shuffle, status interval |
| `writer_settings.toml` | Result buffering, flush interval, channel and batch size, result directory |
| `icmp_settings.toml` | ICMP timeout, retries, workers |
| `tcp_settings.toml` | TCP port, timeout, retries, workers |
| `http_settings.toml` | HTTP/HTTPS/HTTP3 version, TLS range, uTLS fingerprint, accepted status codes |
| `dns_settings.toml` | DNS resolver tuning and DNS tunnel scan orchestration |
| `xray_settings.toml` | Xray connectivity test type, speed test timeout, concurrent dial limits, pre-scan |

You can edit the files by hand or use the in-app inspector. Both write to the same files, and the inspector saves as you go, no restart needed. The [Settings documentation](https://mohsenbg.github.io/bgscan/docs/settings/) explains every field.

---

## Example Usage

**Scan a bundled IP list.** Launch bgscan, pick **Run Scan → IP List → cloudflare_IPv4** (or `cloudflare_IPv6` for an IPv6 sweep), then choose a scan type such as `t` for TCP. bgscan scans the list with the configured workers and writes results to `result/tcp/`, or wherever `result_directory` points.

**Chain ICMP → TCP → HTTP.** Set `pipeline_mode = "streaming"` in `general_settings.toml`. With more than one scan type enabled, the stages chain automatically, and only IPs that pass one stage's success criteria move on to the next.

**Re-scan an earlier result.** Choose **Run Scan → Result List** and pick a saved file. bgscan scans only the IPs in that file, which is handy when you want a closer look at hosts that already passed an earlier stage.

**Validate Xray outbounds.** From the main menu, open **Xray Outbounds** and press `a` to add a template from a share link (`vless://`, `vmess://`, `trojan://`, `ss://`, `hysteria2://`, `wireguard://`) or a JSON file. An Xray scan then tests connectivity and bandwidth.

**Test DNS tunnels.** Open **DNS Tunneling** from the main menu and create a configuration for DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, or TheFeed. Run the DNS Tunneling scan type to see which resolvers can carry the tunnel.

---

## Supported Platforms

| Platform | Architectures |
| :--- | :--- |
| Linux | amd64, arm64, arm32, 386 |
| Windows | amd64, arm64 (10+) |
| macOS | amd64, arm64 |
| Android (Termux) | arm64, arm32, x86_64, x86 |

> Termux: install from F-Droid. The Play Store version is outdated.

---

## Project Structure

```
bgscan/
├── cmd/bgscan/              # Application entry point
├── internal/
│   ├── core/
│   │   ├── config/          # TOML configuration, defaults, and validation
│   │   │   └── validate/   # Per-section and aggregate validators
│   │   ├── scanner/         # Scanner orchestrator and stage builders
│   │   │   ├── engine/      # Streaming and batch pipeline execution
│   │   │   ├── portmgr/     # Local port leasing for tunnel and Xray probes
│   │   │   └── probe/        # Probe interface and per-protocol implementations
│   │   ├── result/          # Result interface, schema, async writer, CSV merge, loader
│   │   ├── iplist/           # IP list loader, parser, registry, shuffle
│   │   ├── netutil/          # netip.Addr helpers and CIDR utilities (IPv4/IPv6)
│   │   ├── dns/              # DNS queries, tunnel config services (DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, TheFeed), SOCKS5, SSH
│   │   ├── speedtest/        # Xray bandwidth, latency, and transport tests
│   │   ├── xray/             # In-process Xray service (vendored core), outbound/link parsing
│   │   ├── process/          # Cross-platform process lifecycle (Slipstream sidecar)
│   │   └── fileutil/         # CSV, JSON, TOML, text, temp-file helpers
│   ├── logger/              # Leveled logging with lumberjack rotation
│   └── ui/                  # BubbleTea TUI (splash, startup, workspace, components, theme)
├── assets/                  # Slipstream sidecar binary, Xray configs and data, DNS tunnel configs + outbound templates
├── ips/                     # Bundled and imported IP lists (CSV)
├── settings/                # Default TOML configuration files
├── result/                  # Scan output (CSV, per scan type)
├── logs/                    # core.log, ui.log, debug.log
├── scripts/                 # Install, build, dependency, publish, and release scripts
├── docs/                    # Hugo Book documentation site
└── go.mod
```

---

## Contributing

Contributions are welcome, whether bug fixes, documentation improvements, or new features.

- Branch from `main` and use a descriptive branch name (`feature/`, `fix/`, `docs/`, `refactor/`).
- Follow existing conventions: keep code simple and readable, and prefer small focused commits.
- Use commit prefixes: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`.
- Before opening a PR, make sure the project builds and the existing tests pass, and update the docs for anything new.
- For large changes, open an Issue first to discuss the approach.

See the full [Contributing Guide](https://mohsenbg.github.io/bgscan/docs/developer/contributing/) in the documentation.

By submitting a contribution to bgscan, you grant permission for your contribution to be distributed as part of the project under the project's license. You confirm that you have the right to submit this contribution and that it does not violate any third-party rights.

---

## License

[GPL-3.0](LICENSE) — Copyright (c) 2026 MohsenBg

---

## Support / Donate

If bgscan has been useful to you, consider supporting its development:

| Network | Currency | Address |
| :---: | :---: | :--- |
| Bitcoin | `BTC` | `bc1qdwh57dm97nmx5jzdr7lrc9cxe5xh3zc59er7z9` |
| Ethereum | `ETH` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| BNB Smart Chain | `BNB / BEP20` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| TRON | `TRX / TRC20` | `TNW6pbfY8zZVZezZWyYXo7h12MycRsVJK7` |

---

<div align="center">

Built with Go · GPL-3.0 License

</div>
