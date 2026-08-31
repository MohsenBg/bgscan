---
title: "Architecture"
weight: 2
---

# Architecture

bgscan is layered. `main` initializes the theme and hands control to a BubbleTea TUI. The TUI runs a three-stage experience — splash, startup, workspace — where the **startup** stage initializes loggers, registers result schemas, loads and validates config, and runs binary health checks. Once startup passes, the workspace drives a multi-stage scanner engine whose probes and writers are pluggable.

## Directory layout

```text
.
├── assets
│   ├── dns-tunneling
│   │   ├── dnstt
│   │   ├── vaydns
│   │   └── slipstream
│   ├── xray
│   │   ├── configs
│   │   └── outbounds
│   └── slipstream-client
├── cmd
│   └── bgscan
│       └── main.go
├── docs
├── internal
│   ├── core
│   │   ├── core.go          schema registration
│   │   ├── config
│   │   │   └── validate
│   │   ├── dns
│   │   ├── fileutil
│   │   ├── iplist
│   │   ├── netutil
│   │   ├── process
│   │   ├── result
│   │   ├── scanner
│   │   │   ├── engine
│   │   │   ├── portmgr
│   │   │   └── probe
│   │   │       ├── dnsttprobe
│   │   │       ├── httpprobe
│   │   │       ├── icmpprobe
│   │   │       ├── resolveprobe
│   │   │       ├── slipstreamprobe
│   │   │       ├── tcpprobe
│   │   │       ├── vaydnsprobe
│   │   │       └── xrayprobe
│   │   ├── socks
│   │   ├── ssh
│   │   ├── speedtest
│   │   └── xray
│   ├── logger
│   └── ui
│       ├── components
│       ├── main
│       │   ├── app
│       │   ├── body
│       │   ├── footer
│       │   ├── header
│       │   ├── splash
│       │   ├── startup
│       │   └── workspace
│       ├── shared
│       └── theme
├── ips
├── scripts
└── settings
```

## Layer overview

```
┌─────────────────────────────────────────────┐
│                 cmd/bgscan                   │
│   theme.Init → app.New → tea.Run             │
├─────────────────────────────────────────────┤
│                internal/ui                   │
│   main                                       │
│     splash  startup  workspace               │
│     app (root model, stage transitions)      │
│     body, header, footer                     │
│   components (menus, tables, inspector)      │
│   shared (layout, env, dialog, ui, valid.)   │
│   theme                                      │
│                                              │
│   ui/main/startup performs (inside the TUI): │
│     logger, core.Init, config, binary checks │
├─────────────────────────────────────────────┤
│               internal/core                  │
│  ┌──────────┐ ┌────────┐ ┌───────────────┐  │
│  │ scanner  │ │ config │ │    result     │  │
│  │ engine   │ │ Store  │ │ writer+schema │  │
│  │ probe    │ │validate│ │   registry    │  │
│  │ portmgr  │ └────────┘ └───────────────┘  │
│  └──────────┘                                │
│  ┌────────┐ ┌─────┐ ┌──────┐ ┌───────────┐  │
│  │ iplist │ │ dns │ │ xray │ │ speedtest │  │
│  └────────┘ └─────┘ └──────┘ └───────────┘  │
│  ┌─────────┐ ┌──────────┐ ┌─────────────┐   │
│  │ netutil │ │ process  │ │  fileutil   │   │
│  └─────────┘ └──────────┘ └─────────────┘   │
├─────────────────────────────────────────────┤
│              internal/logger                 │
├─────────────────────────────────────────────┤
│          assets/  ips/  settings/           │
└─────────────────────────────────────────────┘
```

## Directory reference

| Path | Description |
|---|---|
| `cmd/bgscan` | Entry point. Initializes the theme, constructs the app, and starts the BubbleTea program. |
| `internal/core/core.go` | `core.Init()` registers every built-in probe schema into `result.DefaultRegistry`. Called from the startup stage. |
| `internal/core/config` | `ScannerConfig` types, compiled-in defaults, and the `Store` that reads and writes `settings/*.toml`. |
| `internal/core/config/validate` | Per-section validators and normalizers, combined by `aggregate.go`. |
| `internal/core/scanner` | `Scanner` interface, `StageConfig`, and the stage builders. |
| `internal/core/scanner/engine` | Pipeline execution: single scan, sequential, streaming, batch, pause control. |
| `internal/core/scanner/probe` | `Probe` interface plus one subpackage per probe. |
| `internal/core/scanner/portmgr` | Local port leasing for probes that spawn client binaries. |
| `internal/core/result` | `Result` interface, schemas, registry, async writer, CSV merge, loader. |
| `internal/core/iplist` | IP list import, parsing, registry, shuffle, streaming. |
| `internal/core/netutil` | Host normalization, TLS version parsing, and SNI extraction for the HTTP probes. |
| `internal/core/dns` | DNS queries, transports, DNSTT/VayDNS/Slipstream config services, and tunnel management. |
| `internal/core/socks` | SOCKS5 client for tunnel validation. |
| `internal/core/ssh` | SSH client for tunneled connections through SSH proxies. |
| `internal/core/xray` | Xray process control, inbound and outbound config, share link parsing. |
| `internal/core/speedtest` | Latency, download, and upload measurement used by the Xray probe. |
| `internal/core/process` | Cross-platform process spawn and kill. |
| `internal/core/fileutil` | CSV, JSON, TOML, text, temp-file, sorting, and path helpers. |
| `internal/logger` | Three leveled log streams with lumberjack rotation and live subscribers. |
| `internal/ui/main/app` | Root BubbleTea model. Owns the stage machine: splash → startup → workspace. |
| `internal/ui/main/splash` | Animated splash screen, runs first. |
| `internal/ui/main/startup` | Sequential health checks (Logger, Config, Xray, DNSTT, Slipstream, Vaydns, App) shown to the user with a live status sidebar. |
| `internal/ui/main/workspace` | Main workspace shell (header, body, footer) after startup passes. |
| `internal/ui/main/body` | Stack of body screens (main menu, scan, settings, logs) inside the workspace. |
| `internal/ui/main/header` / `footer` | Workspace chrome. |
| `internal/ui/components` | Widgets, settings inspectors, menus, tables, and the live scanner view. |
| `internal/ui/shared` | Layout geometry, dialog system, key modes, component interface, validation. |
| `internal/ui/theme` | Dark and light palettes plus the huh form theme adapter. |
| `assets/xray` | Bundled Xray binary location, configs, and outbound templates. |
| `assets/dnstt-client`, `assets/slipstream-client` | Optional tunnel client binaries. |
| `assets/dns-tunneling` | DNS tunnel config files (DNSTT, VayDNS, Slipstream). |
| `ips` | Bundled provider IP lists as CSV. |
| `settings` | Live `.toml` settings. |
| `scripts` | Install, build, and release helpers. |

## Application flow

```
main()
  │
  ├─ theme.Init()                   resolve dark / light palette
  │
  ├─ config.AppVersion = Version    recorded for the result files
  │
  ├─ app := app.New()               splash + startup + workspace components
  ├─ p := tea.NewProgram(app)       hand off to BubbleTea
  │
  └─ inside the TUI:
        │
        ├─ StageSplash              animated ASCII logo + version
        │
        ├─ StageStartup             ui/main/startup runs each check
        │     │                     in a goroutine, reports status live
        │     │
        │     ├─ Logger             init loggers, then core.Init()
        │     │                     to register probe schemas
        │     ├─ Config             store.Load(), validate.NormalizeAll,
        │     │                     report any clamped values
        │     ├─ Xray               locate binary, check version
        │     ├─ DNSTT              validate config files
        │     ├─ Slipstream         find binary, verify, validate configs
        │     ├─ Vaydns             validate config files
        │     └─ App                wait for Enter
        │
        └─ StageWorkspace           main workspace:
              ├─ header
              ├─ body               component stack: main menu → scan / settings / logs
              └─ footer

on Run Scan:
  ├─ scanner.NewScanner(ctx, input)
  ├─ AddStage(BuildICMPStage) ...
  ├─ Run()
  │     ├─ one stage   → engine.RunScan
  │     └─ many stages → engine.RunScanWithChain
  │                        sequential | streaming | batch
  └─ result.Writer → CSV merge → disk
```

## Key design principles

**Config is passed, not global.** There is no config singleton. The startup stage constructs a `config.Store`, calls `store.Load()`, and puts the resulting `*ScannerConfig` and `*Store` on `ui.AppState` for components and the scanner to read. Tests construct a `Store` with `WithSettingsDir` against a temp directory, so nothing touches real settings.

**Invalid config self-heals.** `validate.NormalizeAll` clamps out-of-range fields to defaults and reports each correction in the startup sidebar. A bad hand edit degrades to a working default instead of failing the run — corrected values live in memory until the user saves a section through the settings inspector.

**Probes are pluggable.** `probe.Probe` is the only contract: `Init`, `Run`, `Schema`, `Close`. `Run` takes a `netip.Addr`, which is what gives IPv4 and IPv6 one code path.

**Results are self-describing.** Each probe ships a `result.ResultSchema` naming its output directory, columns, and parser. Registering it in `core.Init()` is what makes the writer, the file browser, and the result table understand a new probe. No shared result struct to extend.

**The engine is protocol-agnostic.** It moves addresses and results, and never inspects what a probe measured. Pipeline mode is a config choice, not an engine rewrite.

**External binaries are optional.** ICMP, TCP, HTTP, and DNS resolver probes need only the Go standard library and `golang.org/x/net`. Xray, DNSTT, and Slipstream are checked at startup, and a missing binary disables just that scan type.

**The UI is a component tree.** Every screen implements `ui.Component`. Overlays stack, and the top one consumes all input.

## Related pages

- [Core](../core/) — engine, probes, config, and the result pipeline in detail
- [UI](../ui/) — component model, layout, and theming
- [Getting Started](../getting-started/) — build and run
