---
title: "Architecture"
weight: 2
---

# Architecture

bgscan is a layered application. A lightweight entry point boots startup health checks, then hands control to a BubbleTea TUI. The TUI drives a multi-stage scanner engine whose probes and I/O are pluggable.

---

## Directory layout

```text
.
├── assets
│   └── xray
│       └── outbounds
├── cmd
│   └── bgscan
│       └── main.go
├── docs
├── go.mod
├── internal
│   ├── core
│   │   ├── config
│   │   ├── dns
│   │   ├── fileutil
│   │   ├── ip
│   │   ├── iplist
│   │   ├── process
│   │   ├── result
│   │   ├── scanner
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

---

## Layer overview

```
┌─────────────────────────────────────────────┐
│                  cmd/bgscan                  │
│                  (entry point)               │
├─────────────────────────────────────────────┤
│              internal/startup                │
│          (health checks, config init)        │
├─────────────────────────────────────────────┤
│               internal/ui                    │
│  ┌─────────────────────────────────────┐     │
│  │  main (root model, layout, dialogs)  │     │
│  │  components (menus, tables, forms)   │     │
│  │  shared (layout, env, dialog, ui)    │     │
│  │  theme                               │     │
│  └─────────────────────────────────────┘     │
├─────────────────────────────────────────────┤
│              internal/core                   │
│  ┌──────────────┐ ┌──────────┐ ┌─────────┐  │
│  │  scanner      │ │ config   │ │ result  │  │
│  │  (engine,     │ │ (TOML)   │ │ (writer)│  │
│  │   probe, port) │ │          │ │         │  │
│  └──────────────┘ └──────────┘ └─────────┘  │
│  ┌──────────┐ ┌─────────┐ ┌──────────────┐  │
│  │  iplist  │ │   dns   │ │ xray/process │  │
│  └──────────┘ └─────────┘ └──────────────┘  │
├─────────────────────────────────────────────┤
│              internal/logger                 │
├─────────────────────────────────────────────┤
│            assets/  ips/  settings/         │
│                 (data files)                 │
└─────────────────────────────────────────────┘
```

---

## Directory reference

| Path | Description |
|---|---|
| `cmd/bgscan` | Application entry point (`main.go`). Creates the root TUI model and runs it. |
| `internal/core` | Core logic: config, DNS, file utilities, IP handling, process management, results, scanner, and Xray integration. |
| `internal/core/config` | TOML-based configuration singleton with thread-safe accessors. |
| `internal/core/scanner` | Scanner orchestrator, engine (pipeline), probe interface, port manager, and net utilities. |
| `internal/core/result` | Asynchronous result writer, CSV merge/sort, result file registry, and loader. |
| `internal/core/iplist` | IP list CSV loader, parser, registry, and shuffle. |
| `internal/core/dns` | DNS query, DNSTT, SlipStream, and SOCKS5 helpers. |
| `internal/core/xray` | Xray command runner, inbound/outbound management, and speed test. |
| `internal/core/process` | Cross-platform process lifecycle (spawn, kill, signal). |
| `internal/core/fileutil` | CSV, JSON, TOML, text, temp-file, and path helpers. |
| `internal/core/ip` | IP parsing and CIDR expansion. |
| `internal/logger` | Leveled logging with lumberjack rotation. Three streams: core, ui, debug. |
| `internal/startup` | Sequential startup health checks: logger, theme, config, Xray, DNSTT, Slipstream. |
| `internal/ui/main` | Root BubbleTea model — header/body/footer layout, overlay dialog manager. |
| `internal/ui/components` | Reusable UI: basic widgets, inspector forms, menus, tables, scanner view. |
| `internal/ui/shared` | Shared UI infrastructure: layout geometry, dialog system, env/keys, component interface, validation. |
| `internal/ui/theme` | Dark/light/auto color palettes (Catppuccin-based). |
| `assets/xray/outbounds` | Bundled Xray outbound config assets. |
| `assets/dnstt-client`, `assets/slipstream-client` | DNS tunneling binaries. |
| `ips` | Default IP range lists per provider (Cloudflare, AWS, Azure, etc.) as CSV. |
| `settings` | Default `.toml` settings files. `.default` copies are fallback templates. |
| `scripts` | Install, build, and release helper scripts. |
| `docs` | Hugo Book documentation site. |

---

## Application flow

```
main()
  │
  ├─ startup.RunHealthChecks()
  │     ├─ logger init
  │     ├─ theme init
  │     ├─ config load + validate
  │     ├─ xray binary check
  │     ├─ dnstt binary check
  │     └─ slipstream binary check
  │
  ├─ tea.NewProgram(app.New()).Run()
  │     │
  │     ├─ header  (title bar)
  │     ├─ body    (component stack: main menu → scan/settings/logs/...)
  │     └─ footer (status bar, key hints)
  │
  └─ when user selects Run Scan:
        ├─ scanner.NewScanner(ctx, inputPath)
        ├─ scanner.AddStage(BuildICMPStage)
        ├─ scanner.AddStage(BuildTCPStage)
        ├─ ...
        ├─ scanner.Run()
        │     ├─ single stage → engine.RunScan
        │     └─ multi stage  → engine.RunScanWithChain
        │           ├─ sequential
        │           ├─ streaming
        │           └─ batch
        │
        └─ results written via result.Writer → CSV merge → disk
```

---

## Key design principles

- **No external runtime deps for core scans.** ICMP, TCP, and HTTP probes use Go stdlib. Xray, DNSTT, and Slipstream are optional binaries validated at startup.
- **Configuration is file-first.** All settings live in `settings/*.toml`. The in-app inspector reads and writes the same files.
- **Probes are pluggable.** The `probe.Probe` interface is the only contract. Adding a new scan type means implementing `Init/Run/Close` and registering a stage builder.
- **Engine is pipeline-agnostic.** The engine does not know what the probes do — it just feeds IPs, collects results, and flushes to disk. The pipeline mode (sequential/streaming/batch) is a config choice.
- **UI is a component tree.** Every screen is a `ui.Component` with `Init/Update/View/OnClose/Mode`. Overlays (dialogs, pickers) stack on top; the top overlay consumes all input.

---

## Related pages

- [Core](../core/) — scanner engine, probe interface, config, and result pipeline in detail
- [UI](../ui/) — TUI architecture, component model, layout, and theming
- [Getting Started](../getting-started/) — build and run instructions
