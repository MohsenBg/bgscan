---
title: "Architecture"
weight: 2
---

# Architecture

BGScan has a layered structure. The entry point, `main`, hands control to the Bubble Tea interface. The interface has three stages: Splash, Startup, and Workspace.

During the **Startup** stage, the loggers are checked, the result schemas are registered, the settings are loaded and validated, and the status of the internal Xray core, the tunnel settings files, and the optional Slipstream library is checked.

Once Startup finishes successfully, the app enters the Workspace. From there you can run a scan. The scan engine is built from several stages, and its probes and writers can be replaced.

## Directory layout

```text
.
├── assets
│   ├── dns-tunneling
│   │   ├── dnstt
│   │   ├── masterdns
│   │   ├── slipstream
│   │   ├── stormdns
│   │   ├── thefeed
│   │   └── vaydns
│   ├── xray
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
│   │   ├── ffi
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
│   │   │       ├── masterdnsprobe
│   │   │       ├── resolveprobe
│   │   │       ├── slipstreamprobe
│   │   │       ├── stormdnsprobe
│   │   │       ├── tcpprobe
│   │   │       ├── thefeedprobe
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

```text
┌─────────────────────────────────────────────┐
│                 cmd/bgscan                   │
│   theme.Init → app.New → tea.Run             │
├─────────────────────────────────────────────┤
│                internal/ui                   │
│   main                                       │
│     splash  startup  workspace               │
│     app (root model, stage transitions)      │
│     body, footer                             │
│   components (menus, tables, inspector)      │
│   shared (layout, env, dialog, ui, valid.)   │
│   theme                                      │
│                                              │
│   ui/main/startup performs (inside the TUI): │
│     logger, core.Init, config, shared-lib    │
│     and tunnel-config checks                 │
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
| `cmd/bgscan` | Entry point of the program. It creates the loggers, starts the application (including the theme and the interface settings), and hands control to Bubble Tea. |
| `internal/core/core.go` | The `core.Init()` function registers nine probe schemas in `result.DefaultRegistry`. The MasterDNS and StormDNS schemas are defined but not registered. This function runs during Startup. |
| `internal/core/config` | Defines the `ScannerConfig` types, the compiled defaults, and the `Store` that reads and writes `settings/*.toml`. |
| `internal/core/config/validate` | Validators and normalizers for each settings section, combined in `aggregate.go`. |
| `internal/core/scanner` | The `Scanner` interface, the `StageConfig` struct, and the functions that build stages. |
| `internal/core/scanner/engine` | Runs scan stages in Sequential, Streaming, or Batch mode, along with pause control. Every scan passes through this part. |
| `internal/core/scanner/probe` | The `Probe` interface and one subpackage for each probe type. |
| `internal/core/scanner/portmgr` | Manages leases on local ports for the tunnel and Xray probes. |
| `internal/core/result` | The `Result` interface, schemas, registry, asynchronous writer, CSV merge, and loader. |
| `internal/core/iplist` | Reads and parses IP lists, plus the registry, shuffle, and streaming processing of input. |
| `internal/core/netutil` | Host normalization, TLS version parsing, and SNI extraction for the HTTP probes. |
| `internal/core/dns` | DNS queries, transports, the tunnel settings services (DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed), and tunnel management. |
| `internal/core/socks` | SOCKS5 client for tunnel validation. |
| `internal/core/ssh` | SSH client for tunnel connections through an SSH proxy. |
| `internal/core/xray` | The Xray service that runs in the same process, inbound and outbound settings management, and share link parsing. |
| `internal/core/speedtest` | Measures latency, download speed, and upload speed. Used by the Xray probe. |
| `internal/core/process` | Older process management helpers for several platforms. The current scanner does not use them. |
| `internal/core/ffi` | Connection to the internal Slipstream client library through purego. |
| `internal/core/fileutil` | Helper tools for CSV, JSON, TOML, text, temporary files, sorting, and path handling. |
| `internal/logger` | Two leveled log streams (`core.log` and `ui.log`) with file rotation through lumberjack, plus live subscription to logs. |
| `internal/ui/main/app` | The main Bubble Tea model that manages the Splash, Startup, and Workspace stages. |
| `internal/ui/main/splash` | An animated start screen that shows the ASCII logo and the application version. |
| `internal/ui/main/startup` | Runs the startup checks in order: Logger, Config, Xray, DNSTT, Slipstream, VayDNS, MasterDNS, StormDNS, TheFeed, and App. The status of each check is shown live. |
| `internal/ui/main/workspace` | The main Workspace shell that appears after Startup succeeds and holds the Body and the Footer. |
| `internal/ui/main/body` | Manages the screens inside the Workspace, such as the main menu, scan, settings, and logs. |
| `internal/ui/main/footer` | The status bar at the bottom of the interface. The Workspace has no separate header. |
| `internal/ui/components` | Interface components, settings inspectors, menus, tables, and the live scan view. |
| `internal/ui/shared` | Layout tools, the dialog system, key handling, the component interface, and validation. |
| `internal/ui/theme` | The built in light and dark themes and the theme adapter for huh forms. |
| `assets/xray` | Xray routing data and outbounds. The Xray binary or any generated settings directory is not placed here. Xray runs with the internal core inside the same process. |
| `assets/slipstream-client` | The Slipstream client library, loaded with purego. It does not run as a separate process. |
| `assets/dns-tunneling` | DNS tunnel settings files for DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed. |
| `ips` | Provider IP lists in CSV format. |
| `settings` | The `.toml` settings files that the application uses. |
| `scripts` | Helper scripts for installation, building, and releasing versions. |

## Application flow

```text
main()
  │
  ├─ logger.New("core.log"), logger.New("ui.log")
  │
  ├─ config.AppVersion = Version    recorded for the result files
  │
  ├─ app := app.New()               theme.Init + splash/startup/workspace
  ├─ p := tea.NewProgram(app)       hand off to BubbleTea
  │
  └─ inside the TUI:
        │
        ├─ StageSplash              animated ASCII logo + version
        │
        ├─ StageStartup             ui/main/startup runs each check
        │     │                     in a goroutine, reports status live
        │     │
        │     ├─ Logger             verify loggers, then core.Init()
        │     │                     to register probe schemas
        │     ├─ Config             store.Load(), validate.NormalizeAll,
        │     │                     report any clamped values
        │     ├─ Xray               report embedded core version,
        │     │                     validate outbound templates
        │     ├─ DNSTT              validate config files
        │     ├─ Slipstream         locate and load the client library,
        │     │                     validate configs
        │     ├─ Vaydns             validate config files
        │     ├─ MasterDNS          validate config files
        │     ├─ StormDNS           validate config files
        │     ├─ TheFeed            validate config files
        │     └─ App                wait for Enter
        │
        └─ StageWorkspace           main workspace:
              ├─ body               component stack: main menu → scan / settings / logs
              └─ footer

on Run Scan:
  ├─ scanner.NewScanner(ctx, input)
  ├─ AddStage(BuildICMPStage) ...
  ├─ Run()
  │     └─ engine.RunScanWithChain
  │          sequential | streaming | batch
  └─ result.Writer → CSV merge → disk
```

## Main design decisions

**Settings are passed through structs, not global variables.**

There is no singleton for the settings. The Startup stage creates a `config.Store`, runs `store.Load()`, and makes the `*ScannerConfig` and `*Store` values available to `ui.Session` so components and the scanner can reach them. Tests create a `Store` on a temporary directory with `WithSettingsDir`, so the real settings on the system are not changed.

**Invalid settings are corrected automatically.**

The `validate.NormalizeAll` function returns values that fall outside the allowed range to their default and shows the corrections in the Startup status bar. As a result, a wrong value in a settings file usually gets replaced with a usable default instead of stopping the application. The corrected values stay in memory until the user saves that section through the Inspector, if they want to.

**Probes can be replaced.**

The main probe contract is defined in the `probe.Probe` interface and includes the `Init`, `Run`, `Schema`, and `Close` methods. The `Run` method receives a `netip.Addr`, which is why IPv4 and IPv6 go through one code path.

**The result structure keeps the information it needs for processing.**

Each probe has a `result.ResultSchema` that defines the output directory, the columns, and the parser for its results. The writer uses the same schema directly. Registering schemas in `core.Init()` lets the file browser and the result table connect any file on disk to its own schema. Adding a new result type does not require a shared struct for all probes.

**The scan engine does not depend on protocol details.**

The engine only moves addresses and results between stages and does not care what each probe actually measures. The way scan stages run is chosen through settings, and adding a new mode does not require rewriting the whole engine.

**The scanner does not use external processes.**

The ICMP, TCP, HTTP, and DNS resolver probes need only the Go standard library and `golang.org/x/net`. Xray, DNSTT, VayDNS, MasterDNS, StormDNS, and TheFeed also run through internal libraries inside the same process.

Slipstream works the same way. At startup the application loads the `libslipstream-client` library with purego. If the library is not available, Startup reports the error and Slipstream scans cannot run.

On Termux, the scanner runs the `termux-wake-lock` and `termux-wake-unlock` system commands to keep the CPU awake during a scan. These are the only child processes the scanner starts.

**The interface is built from independent components.**

Every interface screen implements `ui.Component`. Overlays stack on top of each other, and at any moment the topmost overlay receives the user's input.

## Related pages

- [Core](./core/): details of the scan engine, probes, settings, and result processing
- [UI](./ui/): component structure, layout, and themes
- [Getting Started](../getting-started/): guide to building and running the project
