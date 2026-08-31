---
title: "Core"
weight: 3
---

# Core

`internal/core` holds everything that is not the TUI: configuration, the scanner engine, probes, IP lists, results, DNS, Xray, and process management.

## Package map

| Package | Responsibility |
|---|---|
| `config` | `ScannerConfig` types, compiled-in defaults, and the `Store` that reads and writes `settings/*.toml`. |
| `config/validate` | Per-protocol validators and normalizers, aggregated by `aggregate.go`. |
| `scanner` | `Scanner` interface, stage builders, and pipeline assembly. |
| `scanner/engine` | Execution: single scan, sequential chain, streaming pipeline, batch pipeline. |
| `scanner/probe` | `Probe` interface and one subpackage per probe. |
| `scanner/portmgr` | Ephemeral local port pool for probes that spawn tunnel clients. |
| `result` | `Writer`, CSV merge, schema registry, loader, counters. |
| `iplist` | IP list CSV import, parsing, registry, shuffle, streaming. |
| `dns` | DNS query helpers, transport parsing, DNSTT/VayDNS/Slipstream config services, and tunnel management. |
| `socks` | SOCKS5 client for tunnel validation. |
| `ssh` | SSH client for tunneled connections. |
| `xray` | Xray runner, inbound and outbound config, link parsing, speed test. |
| `speedtest` | Latency, download, and upload measurement used by the Xray probe. |
| `netutil` | Host normalization, TLS version parsing, and SNI extraction used by the HTTP probes. |
| `process` | Cross-platform process spawn and kill. |
| `fileutil` | CSV, JSON, TOML, text, temp-file, and path helpers. |

## Scanner

`scanner.Scanner` is an interface, not a struct. The UI holds it, adds stages, and runs it.

```go
type Scanner interface {
    Run() error
    Close() error

    GetStages() []StageConfig
    AddStage(StageConfig)
    UpdateStageHooks(index int, hooks engine.ScanHooks) error

    Pause()
    Resume()
    IsPaused() bool
    PausedDuration() time.Duration

    BuildICMPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildTCPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildHTTPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildXrayStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildResolveStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildDNSTTStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildSlipStreamStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildVayDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
}
```

`BuildXrayStage`, `BuildDNSTTStage`, `BuildSlipStreamStage`, and `BuildVayDNSStage` take a config name and return a slice of stages (an optional resolver pre-scan stage plus the main stage). The other builders return a single stage. All builders accept optional `ScanHooks` via variadic arguments.

A stage is:

```go
type StageConfig struct {
    Workers int
    Probe   probe.Probe
    Writer  result.Writer
    Rate    int
    Hooks   engine.ScanHooks
}
```

Each builder reads its protocol config, constructs the probe, creates a `result.Writer` bound to that probe's schema, and returns the stage. `AddHooks` attaches the UI callbacks afterwards.

```
NewScanner(ctx, input, opts...)
  ├─ AddStage(BuildICMPStage(ctx))
  ├─ AddStage(BuildTCPStage(ctx))
  └─ Run()
      ├─ one stage   → engine.RunScan
      └─ many stages → engine.RunScanWithChain
```

The concrete `scanner` struct takes options: `WithConfig` injects a config instead of loading from disk, `WithPauseController` supplies a shared pause controller. Tests use a `scanRunner` seam to assemble a pipeline without starting workers.

## Engine

`scanner/engine` moves IPs and results. It has no knowledge of what a probe does.

### Single scan

```
iplist.StreamActiveIPs
   │
   ▼
 ips channel ──► worker pool (N goroutines)
                    ├─ rate limiter
                    ├─ probe.Run(ctx, addr)
                    ├─ success → writer + OnSuccess
                    └─ error   → log + OnError
                    │
                    ▼
              result.Writer ──► CSV merge ──► disk

progress goroutine → OnProgress every status_interval
```

`RunScan` returns once the reader finishes, the workers drain, the writer flushes, and a final progress report fires.

### Chain scan

`RunScanWithChain` dispatches on `PipelineMode`:

| Mode | Function | Connection between stages |
|---|---|---|
| `sequential` | `executeSequentialChain` | Stage N writes its file, stage N+1 reads it. |
| `streaming` | `executeStreamingPipeline` | Buffered `chan netip.Addr`, all stages concurrent. |
| `batch` | `executeBatchPipeline` | Fixed-size chunks traverse every stage in turn. |

`createStageChannels` sizes each channel at `MaxBuffer`, falling back to 10,000 when unset, and raises it to the next stage's worker count when that is larger.

`calculateBatchSize` returns `BatchSize` for a single stage, otherwise `max(BatchSize, highest worker count among stages after the first)`, falling back to 1,000 when `BatchSize` is unset.

### Types

```go
type ChainConfig struct {
    Mode             PipelineMode
    MaxBuffer        int
    BatchSize        int
    MaxIPsToTest     uint64
    Stages           []ScanConfig
    MinProbeDuration time.Duration
    Pause            PauseController
    Shuffled         bool
    RateLimiter      *rate.Limiter
}

type ScanConfig struct {
    Workers          int
    MaxIPsToTest     uint64
    MinProbeDuration time.Duration
    ProgressInterval time.Duration
    Probe            probe.Probe
    Writer           result.Writer
    Hooks            ScanHooks
    Pause            PauseController
    Shuffled         bool
    RateLimiter      *rate.Limiter
}

type ScanHooks struct {
    OnProgress func(Progress)
    OnSuccess  func(result.Result)
    OnScanEnd  func()
    OnError    func(error)
}
```

Hooks are optional. The engine goes through `callOnSuccess`, `callOnError`, and `callOnScanEnd`, which no-op on nil.

`ParsePipelineMode` accepts aliases: `simple` for sequential, `parallel` for streaming, `pipeline` for batch. Anything unrecognized returns `ModeSequential`.

### Rate limiting

`GeneralConfig`'s `min_probe_duration`, `probe_per_sec`, and `probe_burst` are converted into a single `rate.Limiter` (token bucket) and a per-probe delay, applied uniformly across every stage in both single and chain scans — they are not per-stage settings. The limiter's `Wait` blocks each worker before `probe.Run`; when a probe returns faster than `MinProbeDuration`, the worker sleeps for the remainder. `probe_per_sec` is the token refill rate (sustained probe ceiling); `probe_burst` is the bucket capacity (max back-to-back probes). `Workers` governs concurrency, not rate — the limiter caps volume independently of worker count.

### Pause control

`PauseController` is an interface implemented by `NewPauseController()`. Workers hit a checkpoint before each probe and block there while paused. `PausedDuration()` accumulates paused time so progress rates stay honest.

## Probe interface

```go
type Probe interface {
    Init(ctx context.Context) error
    Run(ctx context.Context, ip netip.Addr) (result.Result, error)
    Schema() result.ResultSchema
    Close() error
}
```

`Init` runs once before any `Run`, `Close` once at the end. `Run` takes a `netip.Addr`, not a string, which is what gives IPv4 and IPv6 a single code path. `Schema` is what ties a probe to its result layout and output directory.

Each probe lives in its own subpackage under `scanner/probe`:

| Package | Probe | Constructor |
|---|---|---|
| `icmpprobe` | ICMP echo, IPv4 and IPv6 | `NewICMPProbe(Options)` |
| `tcpprobe` | TCP connect | `NewTCPProbe(port, timeout, tries)` |
| `httpprobe` | HTTP/1.1 and HTTP/2 over ALPN | `NewHTTPProbe(req, acceptedCodes)` |
| `httpprobe` | HTTP/3 over QUIC | `NewHTTP3Probe(req, acceptedCodes)` |
| `resolveprobe` | DNS resolver with DPI check | `NewResolverProbe(*DNSRequest)` |
| `dnsttprobe` | DNSTT tunnel validation | `NewDNSTTProbe(config, portMgr)` |
| `vaydnsprobe` | VayDNS tunnel validation | `NewVayDNSProbe(config, portMgr)` |
| `slipstreamprobe` | SlipStream tunnel validation | `NewSlipstreamProbe(workers, config, portMgr)` |
| `xrayprobe` | Xray connectivity and bandwidth | `NewXrayProbe(cfg, template, portMgr)` |

Probes that spawn a client binary take a `portmgr.Manager` and lease a local port per probe so concurrent workers do not collide.

## Result system

Every probe defines a result type implementing `result.Result` and a `result.ResultSchema` describing it.

```go
type Result interface {
    Key() string
    KeyType() KeyType
    ToRecord() []string
    Equal(other Result) bool
    Score() float64
}

type ResultSchema struct {
    Name      string
    Directory string
    Columns   []ColumnDef
    Parser    ResultParser
}
```

`ToRecord` order must match `Columns` order, and `Parser` is the inverse used when loading a file back. `Key` is what deduplication compares, `KeyType` distinguishes IP-keyed from domain-keyed results, and `Score` drives ordering.

### Registration

`core.Init()` in `internal/core/core.go` registers every built-in schema into `result.DefaultRegistry` before anything else runs:

```go
result.DefaultRegistry.Register(icmpprobe.Schema)
result.DefaultRegistry.Register(tcpprobe.Schema)
// http, resolve, dnstt, vaydns, slipstream, xray
```

The registry maps a directory name back to its schema, which is how the result file browser knows how to parse and display a file it finds on disk.

Adding a scan type means writing the result struct, the schema, the probe, and a stage builder, then registering the schema in `core.Init()`. Nothing in the writer, the registry, or the result table needs to change.

### Writer

```go
type Writer interface {
    Start() error
    Stop() error
    Write(r Result)
    GetResultPath() string
}

type WriterOptions struct {
    ResultPrefix string
    Schema       ResultSchema
    Config       config.WriterConfig
}
```

`NewWriter` validates the writer config and the schema, then resolves the output path from `result_directory`, the schema directory, the prefix, and a timestamp. `Start` creates the directory and clears any stale file. `Write` enqueues onto a channel sized by `chan_size` and drops the result if the context is already canceled.

A flush happens when the batch reaches `batch_size`, when `merge_flush_interval` ticks, or on shutdown. On cancellation the writer drains the channel first, so anything accepted before `Stop` reaches disk.

### Merge

`merger.go` sorts the batch by score, merge-sorts it against the existing file, writes to `<path>.tmp`, calls `Sync`, and renames over the original. Duplicates are replaced by key, and neither file is fully loaded into memory.

### Registry and loader

- `registry.go` holds `DefaultRegistry`, mapping a directory name to its schema. `FindResultFiles` and `GetResultFiles` walk the result directory and attach the right schema to each file.
- `loader.go` streams results back through `LoadResult`, or reads a bounded slice with `LoadAll`, both driven by the schema's parser.
- `count.go` counts records without loading the file.

## Config

There is no config singleton and no package-level accessors. `main` builds a `Store`, loads one `ScannerConfig` value, and passes a pointer to it down through the UI and into the scanner.

```go
type ScannerConfig struct {
    General GeneralConfig
    Writer  WriterConfig
    ICMP    ICMPConfig
    TCP     TCPConfig
    HTTP    HTTPConfig
    Xray    XrayConfig
    DNS     DNSConfig
}

store := config.NewStore()          // defaults to the "settings" directory
cfg, err := store.Load()
```

`WithSettingsDir(dir)` points a `Store` somewhere else, which is how tests run against a temp directory instead of real settings.

`Load` reads one TOML file per section. A missing file is created from the compiled-in defaults. A file that exists but does not parse returns an error rather than silently falling back.

Saving is per section: `SaveGeneral`, `SaveWriter`, `SaveICMP`, `SaveTCP`, `SaveHTTP`, `SaveXray`, `SaveDNS`. The inspector edits the in-memory struct and calls the matching method.

The UI carries both through `ui.AppState`:

```go
type AppState struct {
    Layout *layout.Layout
    Config *config.ScannerConfig
    Store  *config.Store
}
```

### Validation

`config/validate` has one file per section. `ValidateXxx` returns a map of field errors and changes nothing. `NormalizeXxx` clamps invalid fields to their defaults and returns a `Warning` for each correction. `aggregate.go` combines both into `ValidateAll` and `NormalizeAll`.

Startup calls `NormalizeAll`, prints every correction, then writes the corrected sections back to disk, so the TOML on disk always matches what the scanner is actually using.

## IP lists

- `parser.go` provides `StreamActiveIPs(ctx, path, limit, shuffled, out)`, the function the engine feeds workers from, and `StreamCIDR` for expanding a prefix.
- `csv.go` and `parser.go` read the two-column format, `<ip_or_cidr>,<enabled>`.
- `loader.go` handles importing external lists, plus `CountIPs` and `CountActiveIPs` for progress totals.
- `registry.go` lists files under `ips/`, `shuffle.go` randomizes order.

Disabled entries are skipped during streaming. Both IPv4 and IPv6 prefixes are supported, so counting a large IPv6 prefix returns a saturated value rather than an exact one.

## DNS subsystem

- `query.go` builds and sends queries over UDP, TCP, and DoT.
- `type.go` parses transports, rcodes, and protocol types. `doh` parses successfully but resolves to DoT, since the scanner targets resolvers by IP.
- `dnstt.go`, `vaydns.go`, and `slipstream.go` define config structs, service interfaces (`DNSTTService`, `VayDNSService`, `SlipstreamService`), and manage tunnel config files under `assets/dns-tunneling/`.
- `shared.go` provides config name normalization, public key validation, and the `GetAllDNSTunsFile` aggregator that merges configs from all three services.
- `socks/` is a SOCKS5 client used to validate a tunnel once it is up.
- `ssh/` is an SSH client for tunneled connections through SSH proxies.

## Xray integration

- `xray.go` and `command.go` spawn and control the process.
- `inbound.go` and `outbound.go` generate the config JSON.
- `link.go` parses share links.
- `speedtest.go` measures throughput.

`xrayprobe` generates a config for the chosen outbound, leases a port, starts Xray, measures latency through the local proxy, optionally runs the speed tests, and tears everything down.

## Process management

`process.go` defines the interface, with `process_unix.go` and `process_windows.go` supplying platform behavior. Used by every probe that spawns a binary.

## Logger

| Logger | File | Covers |
|---|---|---|
| Core | `logs/core.log` | Engine, probes, config, result I/O |
| UI | `logs/ui.log` | Component lifecycle, file operations, UI errors |
| Debug | `logs/debug.log` | State dumps and detailed traces |

Each logger writes to file and fans out to any subscribed viewer channel. Rotation is lumberjack: 50 MB cap, 3 backups, 7 days, compressed.

## Startup

The startup stage lives inside the TUI at `internal/ui/main/startup`. It runs as the second stage of the app model — after the splash animation, before the workspace — and shows each step live in a status sidebar.

The checks are private functions in `checklist.go` driven by the startup component's `Init` and `Update` methods, and they run sequentially in a single goroutine:

```
1. Logger     — initialize core, UI, and debug loggers; call core.Init()
                 to register probe result schemas
2. Config     — store.Load() (fatal on malformed TOML), validate.NormalizeAll
                 to report any out-of-range values clamped to defaults
3. Xray       — locate Xray binary, ensure executable, check version
4. DNSTT      — validate all DNSTT config files
5. Slipstream — find slipstream-client binary, ensure executable, verify,
                 validate config files
6. Vaydns     — validate all VayDNS config files
7. App        — wait for user to press Enter
```

Each check reports status through a `reporter` that emits `[INFO]`, `[SUCCESS]`, `[WARN]`, or `[ERROR]` messages. Critical errors abort subsequent checks. A missing optional binary prints a warning and disables only the scan type that needs it. Once all checks pass and the user presses Enter, the app transitions to the workspace stage.

Unlike the previous design, `main.go` does **not** load config or run any checks — it only calls `theme.Init()`, constructs the app, and starts the BubbleTea program. Config load failures, schema registration, and binary checks all happen as visible steps in the startup UI rather than as silent pre-TUI failures. Corrected values live in memory until the user saves the section through the settings inspector.

## Related pages

- [Architecture](../architecture/) — project layout and layering
- [UI](../ui/) — component model and theming
