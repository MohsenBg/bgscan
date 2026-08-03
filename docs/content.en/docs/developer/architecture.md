---
title: "Architecture"
weight: 2
---

# Architecture

bgscan is layered. `main` registers result schemas, loads config, runs startup checks, then hands control to a BubbleTea TUI. The TUI drives a multi-stage scanner engine whose probes and writers are pluggable.

## Directory layout

```text
.
├── assets
│   ├── dnstt-client
│   ├── slipstream-client
│   └── xray
│       ├── configs
│       └── outbounds
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
│   │   │       └── xrayprobe
│   │   ├── speedtest
│   │   └── xray
│   ├── logger
│   ├── startup
│   └── ui
│       ├── components
│       ├── main
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
│   core.Init → config.Store.Load → startup    │
├─────────────────────────────────────────────┤
│              internal/startup                │
│    logger, theme, config, binary checks      │
├─────────────────────────────────────────────┤
│                internal/ui                   │
│   main (root model, layout, dialog stack)    │
│   components (menus, tables, inspector)      │
│   shared (layout, env, dialog, ui, valid.)   │
│   theme                                      │
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
| `cmd/bgscan` | Entry point. Registers schemas, loads config, runs health checks, starts the TUI. |
| `internal/core/core.go` | `core.Init()` registers every built-in probe schema into `result.DefaultRegistry`. |
| `internal/core/config` | `ScannerConfig` types, compiled-in defaults, and the `Store` that reads and writes `settings/*.toml`. |
| `internal/core/config/validate` | Per-section validators and normalizers, combined by `aggregate.go`. |
| `internal/core/scanner` | `Scanner` interface, `StageConfig`, and the stage builders. |
| `internal/core/scanner/engine` | Pipeline execution: single scan, sequential, streaming, batch, pause control. |
| `internal/core/scanner/probe` | `Probe` interface plus one subpackage per probe. |
| `internal/core/scanner/portmgr` | Local port leasing for probes that spawn client binaries. |
| `internal/core/result` | `Result` interface, schemas, registry, async writer, CSV merge, loader. |
| `internal/core/iplist` | IP list import, parsing, registry, shuffle, streaming. |
| `internal/core/netutil` | Host normalization, TLS version parsing, and SNI extraction for the HTTP probes. |
| `internal/core/dns` | DNS queries, transports, DNSTT and SlipStream clients, SOCKS5. |
| `internal/core/xray` | Xray process control, inbound and outbound config, share link parsing. |
| `internal/core/speedtest` | Latency, download, and upload measurement used by the Xray probe. |
| `internal/core/process` | Cross-platform process spawn and kill. |
| `internal/core/fileutil` | CSV, JSON, TOML, text, temp-file, sorting, and path helpers. |
| `internal/logger` | Three leveled log streams with lumberjack rotation and live subscribers. |
| `internal/startup` | Health checks: logger, theme, config, Xray, DNSTT, SlipStream. |
| `internal/ui/main` | Root model with header, body, footer, and the overlay dialog stack. |
| `internal/ui/components` | Widgets, settings inspectors, menus, tables, and the live scanner view. |
| `internal/ui/shared` | Layout geometry, dialog system, key modes, component interface, validation. |
| `internal/ui/theme` | Dark and light palettes plus the huh form theme adapter. |
| `assets/xray` | Bundled Xray binary location, configs, and outbound templates. |
| `assets/dnstt-client`, `assets/slipstream-client` | Optional tunnel client binaries. |
| `ips` | Bundled provider IP lists as CSV. |
| `settings` | Live `.toml` settings. The `.toml.default` copies are reference snapshots and are not read at runtime. |
| `scripts` | Install, build, and release helpers. |

## Application flow

```
main()
  │
  ├─ core.Init()                    register probe result schemas
  │
  ├─ config.NewStore()              settings/ directory
  ├─ store.Load()                   create missing files from defaults,
  │                                 error on malformed TOML
  │
  ├─ startup.RunHealthChecks(&cfg, &store)
  │     ├─ checkLoggerHealth()
  │     ├─ theme.Init()
  │     ├─ checkConfigHealth()      normalize, then write corrections back
  │     ├─ checkXrayHealth()
  │     ├─ checkDNSTTHealth()
  │     └─ checkSlipstreamHealth()
  │
  ├─ tea.NewProgram(app.New(&cfg, &store)).Run()
  │     ├─ header
  │     ├─ body     component stack: main menu → scan / settings / logs
  │     └─ footer
  │
  └─ on Run Scan:
        ├─ scanner.NewScanner(ctx, input)
        ├─ AddStage(BuildICMPStage) ...
        ├─ Run()
        │     ├─ one stage   → engine.RunScan
        │     └─ many stages → engine.RunScanWithChain
        │                        sequential | streaming | batch
        └─ result.Writer → CSV merge → disk
```

## Key design principles

**Config is passed, not global.** There is no config singleton. `config.Store` loads a `ScannerConfig` value in `main`, and that pointer travels through `ui.AppState` to every component and into the scanner. Tests construct a `Store` with `WithSettingsDir` against a temp directory, so nothing touches real settings.

**Invalid config self-heals.** `validate.NormalizeAll` clamps out-of-range fields to defaults, reports each correction, and startup writes the corrected sections back to disk. A bad hand edit degrades to a working default instead of failing the run.

**Probes are pluggable.** `probe.Probe` is the only contract: `Init`, `Run`, `Schema`, `Close`. `Run` takes a `netip.Addr`, which is what gives IPv4 and IPv6 one code path.

**Results are self-describing.** Each probe ships a `result.ResultSchema` naming its output directory, columns, and parser. Registering it in `core.Init()` is what makes the writer, the file browser, and the result table understand a new probe. No shared result struct to extend.

**The engine is protocol-agnostic.** It moves addresses and results, and never inspects what a probe measured. Pipeline mode is a config choice, not an engine rewrite.

**External binaries are optional.** ICMP, TCP, HTTP, and DNS resolver probes need only the Go standard library and `golang.org/x/net`. Xray, DNSTT, and SlipStream are checked at startup, and a missing binary disables just that scan type.

**The UI is a component tree.** Every screen implements `ui.Component`. Overlays stack, and the top one consumes all input.

## Related pages

- [Core](../core/) — engine, probes, config, and the result pipeline in detail
- [UI](../ui/) — component model, layout, and theming
- [Getting Started](../getting-started/) — build and run
