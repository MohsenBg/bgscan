---
title: "Core"
weight: 3
---

# Core

The `internal/core` package contains all non-UI logic: configuration, the scanner engine, probe implementations, IP list handling, results, DNS, Xray integration, and process management.

---

## Package map

| Package | Responsibility |
|---|---|
| `config` | Thread-safe singleton holding all TOML settings; load/save helpers; validation entry points. |
| `config/validate` | Per-protocol validators (ICMP, TCP, HTTP, DNS, Xray, writer, general). |
| `scanner` | `Scanner` orchestrator and stage builders. |
| `scanner/engine` | Pipeline execution: single scan, sequential chain, streaming pipeline, batch pipeline. |
| `scanner/probe` | `Probe` interface and all concrete probes (ICMP, TCP, HTTP, HTTP/3, DNS, DNSTT, SlipStream, Resolve, Xray). |
| `scanner/portmgr` | Ephemeral port pool for probes that need local bind ports (Xray, DNSTT, SlipStream). |
| `scanner/netutil` | Shared network utilities. |
| `result` | `Writer` (async batch+merge), CSV format, loader, registry, ordering. |
| `iplist` | CSV IP list loader, parser, registry, shuffle, streaming. |
| `ip` | IPv4 parsing and CIDR expansion. |
| `dns` | DNS query helpers, transport parsing, DNSTT/SlipStream SOCKS5 client. |
| `xray` | Xray binary runner, inbound/outbound config, link parsing, speed test. |
| `process` | Cross-platform process spawn/kill (Unix + Windows split). |
| `fileutil` | CSV, JSON, TOML, text, temp-file, and path helpers. |
| `logger` | Leveled logging (core, ui, debug) with lumberjack rotation. |

---

## Scanner

`internal/core/scanner/scanner.go` defines the `Scanner` struct — the public API the UI calls to run a scan.

```go
type Scanner struct {
    ctx    context.Context
    cancel context.CancelFunc
    pause  *engine.PauseController
    input  string
    pm     *portmgr.PortManager
    stages []StageConfig
}
```

`StageConfig` bundles a `ScanMode`, worker count, `probe.Probe`, `result.Writer`, rate limit, and optional `ScanHooks`.

**Lifecycle:**

```
NewScanner(ctx, input)
  ├─ AddStage(BuildICMPStage(ctx))
  ├─ AddStage(BuildTCPStage(ctx))
  └─ Run()
      ├─ single stage → engine.RunScan(ctx, input, maxIPs, cfg, shuffled, pause)
      └─ multi stage  → engine.RunScanWithChain(ctx, input, maxIPs, chainCfg)
```

`Scanner` also exposes `Pause()`, `Resume()`, `IsPaused()`, `PausedDuration()`, and `Close()` for UI control.

**Stage builders** (`BuildICMPStage`, `BuildTCPStage`, `BuildHTTPStage`, `BuildXrayStage`, `BuildResolveStage`, `BuildDNSTTStage`, `BuildSlipStreamStage`) each:

1. Read protocol config via `config.GetXxx()`.
2. Build a result file path and create a `result.Writer`.
3. Construct the appropriate `probe.Probe`.
4. Return a `StageConfig` with workers, rate, and writer.

---

## Engine

`internal/core/scanner/engine` is the execution core. It does not know what probes do — it only moves IPs and results.

#### Single scan: `RunScan`

`engine/scan.go` orchestrates a standalone stage:

```
Reader (iplist.StreamActiveIPs)
   │
   ▼
  ips channel ──► Worker Pool (N goroutines)
                      │
                      ├─ rate limiter (rateCh)
                      ├─ probe.Run(ctx, ip)
                      ├─ on success: results channel + OnSuccess hook
                      └─ on error: logger + OnError hook
                      │
                      ▼
                 results channel ──► Writer goroutine ──► disk (CSV merge)

Progress reporter goroutine → OnProgress hook (interval = status_interval)
```

`RunScan` blocks until the reader finishes, workers drain, writer flushes, and a final progress report fires.

#### Chain scan: `RunScanWithChain`

`engine/chain.go` dispatches based on `PipelineMode`:

| Mode | Function | How stages connect |
|---|---|---|
| `sequential` | `executeSequentialChain` | Stage N writes to disk; stage N+1 reads that file as input. |
| `streaming` | `executeStreamingPipeline` | Buffered channels between stages. All stages run concurrently. |
| `batch` | `executeBatchPipeline` | IPs chunked into batches; each batch traverses all stages before next. |

**Sequential** — lowest memory, slowest. Each stage fully completes before the next starts. The writer's result path becomes the next stage's input.

**Streaming** — highest throughput. `createStageChannels` builds buffered `chan string` between stages. Channel size = `max(workers, MaxIPsPerStage)`. Successful IPs flow instantly to the next stage via `output <- ip`.

**Batch** — hybrid. `streamIPsFromFile` chunks IPs into `batchSize` slices. `processBatch` runs each slice through all `stageExecutor`s sequentially. The next batch doesn't start until the current one finishes all stages.

#### Types

```go
type ChainConfig struct {
    Mode      PipelineMode    // sequential | streaming | batch
    MaxBuffer int             // channel buffer size between streaming stages
    Stages    []ScanConfig
    Pause     *PauseController
    Shuffled  bool
}

type ScanConfig struct {
    Workers int
    Rate    int
    Probe   probe.Probe
    Writer  *result.Writer
    Hooks   ScanHooks
}

type ScanHooks struct {
    OnProgress func(Progress)
    OnSuccess  func(result.IPScanResult)
    OnScanEnd  func()
    OnError    func(error)
}
```

All hooks are optional (nil = disabled). The engine calls `callOnSuccess`, `callOnError`, `callOnScanEnd` safely.

#### Pause control

`engine/pause.go` provides `PauseController` — a non-blocking pause/resume mechanism. Workers check `pause.IsPaused()` and skip work while paused. `PausedDuration()` tracks total paused time so progress reporting stays accurate.

---

## Probe interface

`scanner/probe/probe.go` defines the single contract all probes implement:

```go
type Probe interface {
    Init(ctx context.Context) error
    Run(ctx context.Context, ip string) (*result.IPScanResult, error)
    Close() error
}
```

- **`Init`** — called once at startup. Allocate sockets, spawn goroutines, open caches.
- **`Run`** — called per IP. Must honor `ctx` for cancellation. Returns `IPScanResult` on success, error on failure/timeout.
- **`Close`** — called once at shutdown. Release sockets, goroutines, file descriptors.

#### Available probes

| File | Probe | Constructor |
|---|---|---|
| `icmp.go` | ICMP echo | `NewICMPProbe(timeout, tries)` |
| `tcp.go` | TCP connect | `NewTCPProbe(port, timeout, tries)` |
| `http.go` | HTTP/1.1 + HTTP/2 (ALPN) | `NewHTTPProbe(reqCfg, acceptedCodes)` |
| `http3.go` | HTTP/3 over QUIC | `NewHTTP3Probe(reqCfg, acceptedCodes)` |
| `httpshare.go` | Shared HTTP request config builder | `NewHTTPRequestFromConfig` / `NewHTTP3RequestFromConfig` |
| `resolve.go` | DNS resolver (A/AAAA, DPI check) | `NewResolverProbe(DnsRequest)` |
| `dnstt.go` | DNSTT tunnel validation | `NewDNSTTProbe(DNSTTConfig, portMgr)` |
| `slipstream.go` | SlipStream SOCKS validation | `NewSlipstreamProbe(workers, SlipstreamConfig, portMgr)` |
| `xray.go` | Xray outbound connectivity + bandwidth | `NewXrayProbe(cfg, template, portMgr)` |
| `processes.go` | Xray/DNSTT/Slipstream process lifecycle | used by probes that spawn binaries |

Probes that need local bind ports (Xray, DNSTT, SlipStream) receive a `*portmgr.PortManager` — an ephemeral port pool that allocates and recycles ports to avoid collisions.

---

## Config

`internal/core/config/config.go` exposes a thread-safe singleton:

```go
func Get() *ScannerConfig          // root singleton
func GetGeneral() *GeneralConfig   // per-protocol accessors (RLock/Unlock)
func GetTCP() *TCPConfig
// ... ICMP, HTTP, Xray, DNS, Writer
```

**File layout** — each protocol has its own TOML in `settings/`:

```
settings/
├── general_settings.toml
├── writer_settings.toml
├── icmp_settings.toml
├── tcp_settings.toml
├── http_settings.toml
├── xray_settings.toml
└── dns_settings.toml
```

**Load flow:**

1. `config.Init()` calls all `LoadXxxConfig()` functions in sequence.
2. Each loader reads the TOML file or falls back to defaults (`.default` copy).
3. The loaded struct replaces the singleton field via a private setter (write-locked).
4. Immediately after `Init()`, `startup.checkConfigHealth()` runs validators from `config/validate/` to normalize values.

**Save flow:**

1. UI inspector calls `SaveXxxConfig(cfg)`.
2. `saveConfig` writes TOML to disk via `fileutil.WriteTOMLFile`.
3. The in-memory singleton field is updated via the setter.

#### Validators

`config/validate/` contains one validator per protocol (`validate_icmp.go`, `validate_tcp.go`, etc.) — all called from `validate/all.go`. They clamp values to safe bounds and surface configuration errors at startup before a scan ever runs.

---

## Result pipeline

`internal/core/result/` handles writing scan output to disk.

#### Writer

`result.Writer` is an asynchronous batch-and-merge writer:

```go
type Writer struct {
    config    Config
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    resultPath string
    input     chan IPScanResult
    batch     []IPScanResult
    batchSize int
}
```

**Flush triggers:**

- `BatchSize` results accumulated
- `MergeFlushInterval` elapses
- `Stop()` called (shutdown flush)

The writer guarantees every result written to its channel before `Stop()` is flushed to disk.

#### Merge

`result/merger.go` implements a crash-safe streaming merge:

- Sorts the batch, merge-sorts it against the existing file.
- Writes to `resultPath.tmp`, `fsync`s, then atomically renames to the final path.
- Duplicate IPs are replaced by the newer record.
- Constant memory — never loads both files fully into RAM.

#### Result format

Each result file is a CSV with columns matching `IPScanResult`:

```go
type IPScanResult struct {
    IP       string
    Latency  time.Duration
    Download time.Duration
    Upload   time.Duration
}
```

Files are organized by scan type:

```
results/
├── icmp/
├── tcp/
├── http/
├── xray/
├── dnstt/
├── slipstream/
└── resolve/
```

#### Registry and loader

- `registry.go` — discovers and classifies result files by type from the result directory.
- `loader.go` — streams results back from CSV (used when re-scanning from a result list).
- `count.go` — counts records without loading the full file into memory.

---

## IP lists

`internal/core/iplist/` handles input file loading:

- `loader.go` — `StreamActiveIPs(ctx, path, maxIP, shuffled, out)` — the function the engine calls to feed the worker pool.
- `parser.go` / `csv.go` — parses the internal 2-column CSV format (`<ip_or_cidr>,<enable>`).
- `registry.go` — lists and manages IP files in `ips/`.
- `shuffle.go` — randomizes target order.
- `ip/expand.go` — expands CIDR ranges to individual IPs.

Disabled entries (`enable=0`) are silently skipped during streaming.

---

## DNS subsystem

`internal/core/dns/` provides shared DNS helpers used by the resolver, DNSTT, and SlipStream probes:

- `query.go` — low-level DNS query construction and sending (UDP, TCP, DNS-over-TLS).
- `type.go` — DNS type and rcode parsing helpers (`ParseDNSRcode`, `ParseTransport`).
- `dnstt.go` / `slipstream.go` — tunnel client wrappers.
- `socks5.go` — minimal SOCKS5 client for tunnel validation.

The DNS resolver probe supports DPI checks, EDNS0 buffer sizes, random subdomains, and configurable query types.

---

## Xray integration

`internal/core/xray/` manages Xray binary interaction:

- `xray.go` / `command.go` — spawns and controls the Xray process.
- `inbound.go` / `outbound.go` — generates Xray config JSON for inbound/outbound.
- `link.go` — parses Xray share links (vless://, vmess://, etc.).
- `speedtest.go` — upload/download bandwidth measurement.

The Xray probe (`probe/xray.go`) ties these together: it generates a config, spawns Xray, tests connectivity, optionally runs a speed test, then tears down.

---

## Process management

`internal/core/process/` abstracts cross-platform process lifecycle:

- `process.go` — shared interface.
- `process_unix.go` — Unix-specific spawn/kill (setsid, signal handling).
- `process_windows.go` — Windows-specific (CREATE_NEW_PROCESS_GROUP, taskkill).

Used by probes that spawn external binaries (Xray, DNSTT, SlipStream).

---

## logger

`internal/logger/` provides three leveled streams:

| Logger | File | What it covers |
|---|---|---|
| Core | `logs/core.log` | Scanner engine, probes, config, result I/O |
| UI | `logs/ui.log` | Component lifecycle, file ops, UI-level errors |
| Debug | `logs/debug.log` | Verbose state dumps, goroutine dumps, detailed traces |

Log files rotate via lumberjack: 50 MB cap, 3 backups, 7-day retention, gzip-compressed.

---

## Startup

`internal/startup/health.go` runs `RunHealthChecks()` before the TUI starts:

```
1. checkLoggerHealth()    → open log files, init rotation
2. theme.Init()           → resolve dark/light palette
3. checkConfigHealth()    → config.Init() + validate/all.go
4. checkXrayHealth()      → locate Xray binary + templates
5. checkDNSTTHealth()     → locate DNSTT client binary
6. checkSlipstreamHealth() → locate SlipStream client binary
```

Missing binaries (Xray/DNSTT/SlipStream) log a warning and disable the dependent scan type — the app still runs. Config validation issues are fatal.

> 💡 Set `fastboot = true` in `health.go` to skip the 500 ms pauses between checks during debugging.

---

## Related pages

- [Architecture](../architecture/) — high-level project layout
- [UI](../ui/) — TUI architecture and component model
