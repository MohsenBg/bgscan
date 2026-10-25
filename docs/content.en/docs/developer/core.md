---
title: "Core"
weight: 3
---

# Core

The `internal/core` directory holds everything in BGScan except the TUI interface. That includes configuration and the scan engine, along with probes, IP lists, results, DNS, Xray, and the old process management tool.

## Package map

| Package | Responsibility |
|---|---|
| `config` | Defines `ScannerConfig`, the compiled-in defaults, and the `Store` that reads and writes `settings/*.toml`. |
| `config/validate` | Validates and normalizes the settings for each protocol. `aggregate.go` combines these parts. |
| `scanner` | The `Scanner` interface, stage builders, and the assembly of scan stages. |
| `scanner/engine` | Runs scans in Sequential, Streaming, or Batch mode. Sequential mode is also used for a single stage scan. |
| `scanner/probe` | The `Probe` interface and the subpackage for each probe. |
| `scanner/portmgr` | Manages the pool of temporary local ports for tunnel and Xray probes. |
| `result` | The `Writer`, CSV file merging, the schema registry, the loader, and result counters. |
| `iplist` | Reads IP lists from CSV, parses them, keeps the registry, shuffles entries, and streams inputs. |
| `dns` | DNS query helpers, transport parsing, configuration management for the DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed tunnels, and tunnel management. |
| `socks` | A SOCKS5 client used to validate a tunnel. |
| `ssh` | An SSH client for tunneled connections. |
| `xray` | The in-process Xray service, inbound and outbound configuration, link parsing, and speed tests. |
| `speedtest` | Measures latency along with download and upload speed. The Xray probe uses it. |
| `netutil` | Normalizes hosts, parses TLS versions, and extracts the SNI for HTTP probes. |
| `process` | The old process tracking tool. The current scanner and TUI do not use it. |
| `ffi` | Connects to the Slipstream client library through purego. |
| `fileutil` | Helpers for CSV, JSON, TOML, text, temporary files, and paths. |

## Scanner

`scanner.Scanner` is an interface rather than a struct. The interface keeps an instance of it, adds stages to it, and runs the scan.

```go
type Scanner interface {
    Run() error
    Close() error

    GetStages() []StageConfig
    AddStage(StageConfig) error
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
    BuildMasterDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildStormDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildTheFeedStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
}
```

The methods `BuildXrayStage`, `BuildDNSTTStage`, `BuildSlipStreamStage`, `BuildVayDNSStage`, `BuildMasterDNSStage`, `BuildStormDNSStage`, and `BuildTheFeedStage` take a config file name and return a slice of stages. That slice can contain an optional pre-scan stage and the main stage.

Xray uses the `pre_scan_type` option, which can be `tcp`, `icmp`, `http`, or `none`. The DNS tunnel builders also add a resolver check stage when `check_dns_resolver` is enabled. The remaining builders return only one stage.

Every builder can receive `ScanHooks` as variadic arguments, so a separate signature is not needed to use hooks.

A stage has this structure:

```go
type StageConfig struct {
    Workers          int
    ProgressInterval time.Duration
    Probe            probe.Probe
    Writer           result.Writer
    Hooks            engine.ScanHooks
}
```

Each builder reads the settings for its own protocol, constructs the probe, creates a `result.Writer` that matches the probe schema, and returns the stage. The `AddHooks` method attaches the interface callbacks.

The general flow for using the Scanner is:

```text
NewScanner(ctx, input, opts...)
 ├─ AddStage(BuildICMPStage(ctx))
 ├─ AddStage(BuildTCPStage(ctx))
 └─ Run()
     └─ engine.RunScanWithChain
          sequential | streaming | batch
```

The actual `scanner` implementation takes several options. The `WithConfig` option lets the caller supply a config from outside instead of having the scanner read it from disk. `WithPauseController` gives it a shared pause controller.

Tests use the `scanRunner` seam to assemble scan stages without starting workers.

## Engine

The `scanner/engine` package runs the scan stages. It deals only with IP addresses and results and knows nothing about what each probe does.

### Single stage scan

A single stage scan uses the same execution path as any other scan. It only runs in Sequential mode. The Engine has no separate path for a single stage scan.

```text
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

A scan finishes when the reader reaches the end of the input, every worker has finished its work, the Writer has flushed its data, and the final progress report has been sent.

### Chain scan

The `RunScanWithChain` function chooses how to run the stages based on the value of `PipelineMode`.

| Mode | Function | How stages connect |
|---|---|---|
| `sequential` | `runSequential` | Each stage writes its results to a file and the next stage reads that file. |
| `streaming` | `runStreaming` | All stages run at the same time and IPs move between them through a buffered `chan netip.Addr`. |
| `batch` | `runBatch` | Data is split into fixed size batches and each batch passes through the stages in order. |

The `createStageChannels` function takes each channel buffer size from `MaxBuffer`. When that value is not set, the default is `10,000`. If the next stage has more workers, the buffer size grows to match that count.

The `pipeline.batchSize` function returns `BatchSize` for a single stage scan. In a multi stage scan it picks the larger value between `BatchSize` and the highest worker count among the stages after the first stage. When `BatchSize` is not set, the default is `1,000`.

### Structures and types

```go
type ChainConfig struct {
    Mode             PipelineMode
    MaxBuffer        int
    BatchSize        int
    MaxIPsToTest     uint64
    Stages           []StageConfig
    MinProbeDuration time.Duration
    Pause            PauseController
    Shuffled         bool
    RateLimiter      *rate.Limiter
    MaxSuccessfulIPs uint64
    Log              *logger.Logger
}

type StageConfig struct {
    Workers          int
    ProgressInterval time.Duration
    Probe            probe.Probe
    Writer           result.Writer
    Hooks            ScanHooks
}

type ScanHooks struct {
    OnProgress func(Progress)
    OnSuccess  func(result.Result)
    OnScanEnd  func()
    OnError    func(error)
}
```

Hooks are optional. The Engine uses the `callOnSuccess`, `callOnError`, and `callOnScanEnd` functions to run callbacks, and these do nothing when the matching callback is `nil`.

The `ParsePipelineMode` function recognizes only the values `sequential`, `streaming`, and `batch`. Values such as `simple`, `parallel`, and `pipeline` are treated as unknown modes and the result is `ModeSequential`.

The settings validator still accepts the three old names, but the Engine does not treat them as known modes. They pass validation, yet execution ends up in Sequential mode.

### Rate limiting

The `min_probe_duration`, `probe_per_sec`, and `probe_burst` options from `GeneralConfig` become a token bucket `rate.Limiter` and a delay for each probe. These limits apply to every stage in both single stage and chain scans, and there is no separate setting for each stage.

- `probe_per_sec` sets the rate at which the token bucket refills and affects the sustained rate of probes.
- `probe_burst` sets the bucket capacity, meaning how many probes can run back to back.
- `Workers` controls how many jobs run at once, not the probe rate.
- The limiter caps the rate independently of the worker count.

Each worker waits on the limiter's `Wait` before running `probe.Run`. If a probe finishes sooner than `MinProbeDuration`, the worker waits for the remaining time.

### Pause control

The `PauseController` interface defines how the scan pauses and resumes, and the `NewPauseController()` function builds its implementation.

Workers reach a checkpoint before running each probe. If the scan is paused, they wait there until execution resumes.

The `PausedDuration()` method measures how long the scan has been paused so that stopped time does not distort the real progress rate.

## Probe interface

```go
type Probe interface {
    Init(ctx context.Context) error
    Run(ctx context.Context, ip netip.Addr) (result.Result, error)
    Schema() result.ResultSchema
    Close() error
}
```

The `Init` method is called once before the `Run` calls and `Close` runs at the end. The `Run` method receives a `netip.Addr` instead of a string, which is why IPv4 and IPv6 share one code path.

The `Schema` method defines how a probe's results are stored and which directory they go into.

Each probe is implemented in its own subpackage under `scanner/probe`:

| Package | Probe type | Constructor |
|---|---|---|
| `icmpprobe` | Checks the ICMP response over IPv4 and IPv6 | `NewICMPProbe(Options)` |
| `tcpprobe` | Checks a TCP connection | `NewTCPProbe(port, timeout, tries)` |
| `httpprobe` | Checks HTTP/1.1 and HTTP/2 through ALPN | `NewHTTPProbe(req, acceptedCodes)` |
| `httpprobe` | Checks HTTP/3 over QUIC | `NewHTTP3Probe(req, acceptedCodes)` |
| `resolveprobe` | Checks a DNS resolver with DPI detection | `NewResolverProbe(*DNSRequest)` |
| `dnsttprobe` | Validates a DNSTT tunnel | `NewDNSTTProbe(config, timeout, ...)` |
| `vaydnsprobe` | Validates a VayDNS tunnel | `NewVayDNSProbe(config, timeout, ...)` |
| `slipstreamprobe` | Validates a Slipstream tunnel | `NewSlipstreamProbe(config, timeout, portMgr, ...)` |
| `masterdnsprobe` | Validates a MasterDNS tunnel | `NewMasterDNSProbe(config, timeout, portMgr, ...)` |
| `stormdnsprobe` | Validates a StormDNS tunnel | `NewStormDNSProbe(config, timeout, portMgr, ...)` |
| `thefeedprobe` | Validates a TheFeed tunnel | `NewTheFeedProbe(config, timeout, ...)` |
| `xrayprobe` | Checks Xray connectivity and bandwidth | `NewXrayProbe(cfg, template, portMgr)` |

Only the Slipstream probe loads an internal native library, named `libslipstream-client`.

The Slipstream, MasterDNS, StormDNS, and Xray probes receive a `portmgr.Manager` and lease a local port for each probe so that concurrent workers do not share one port.

## Result system

Each probe has its own result type, which implements the `result.Result` interface. Each probe also has a `result.ResultSchema` that defines the result structure.

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

The order of the values returned by `ToRecord` must match the order of `Columns`. The `Parser` turns stored records back into results when a file is read again.

- `Key` defines how results are compared to find duplicates.
- `KeyType` separates results keyed by IP from results keyed by domain.
- `Score` is used to sort results.

### Registering schemas

The `core.Init()` function in `internal/core/core.go` registers nine built-in schemas in `result.DefaultRegistry`:

```go
result.DefaultRegistry.Register(icmpprobe.Schema)
result.DefaultRegistry.Register(tcpprobe.Schema)
result.DefaultRegistry.Register(httpprobe.Schema)
result.DefaultRegistry.Register(resolveprobe.Schema)
result.DefaultRegistry.Register(dnsttprobe.Schema)
result.DefaultRegistry.Register(vaydnsprobe.Schema)
result.DefaultRegistry.Register(slipstreamprobe.Schema)
result.DefaultRegistry.Register(thefeedprobe.Schema)
result.DefaultRegistry.Register(xrayprobe.Schema)
```

The `masterdnsprobe` and `stormdnsprobe` schemas are defined but do not appear in this list. As a result, the registry cannot connect the directories for these two protocols to their schemas, and their files do not show up in the Result Files browser.

Even so, the results from these two protocols are still saved as CSV under `result/masterdns/` and `result/stormdns/`.

The registry connects a directory name to its schema. This is how the result file browser knows how to read and display each file on disk.

To add a new scan type, implement its result type, schema, probe, and stage builder, then register the schema in `core.Init()`. This does not require changes to the Writer, the registry, or the result table.

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
    Log          *logger.Logger
}
```

The `NewWriter` function validates the writer settings and the schema. It then builds the output path from `result_directory`, the schema directory, the file prefix, and a timestamp.

The `Start` method creates the output directory and deletes old files. The `Write` method puts the result into a channel with the capacity given by `chan_size`. If the context has already been canceled, the result is dropped.

The Writer writes data to disk in three situations:

- When the number of records in a batch reaches `batch_size`.
- When the `merge_flush_interval` time arrives.
- When the Writer shuts down.

On cancellation the Writer first processes the data still left in the channel. This gives data accepted before the `Stop` call a chance to reach disk.

### Merging results

The `merger.go` file sorts and merges results. It first sorts the batch by score and then merges it with the existing file.

The result is written to the temporary file `\<path>.tmp`, then `Sync` runs, and finally the temporary file replaces the original.

Duplicate records are compared by key and only the better record is kept. Merging also does not require loading the full contents of both files into memory at the same time.

### Registry and loader

- `registry.go` holds `DefaultRegistry`, which connects a directory name to its schema. The `FindResultFiles` and `GetResultFiles` functions walk the result directory and attach the right schema to each file.
- `loader.go` reads results as a stream with `LoadResult` and builds a bounded slice of results with `LoadAll`. Both use the parser from the schema.
- `count.go` counts records without loading all results into memory.

## Settings

Config does not use a singleton or a global accessor at package level. The Startup checks create a `Store`, read `ScannerConfig` settings, and hand them to the UI Session so that the interface and the scanner use the same settings.

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

store := config.NewStore() // defaults to the "settings" directory
cfg, err := store.Load()
```

The `WithSettingsDir(dir)` option changes the settings path for a `Store`. Tests use it to read and write settings in a temporary directory instead of the real system settings path.

The `Load` method reads the TOML file for each section. If a file does not exist, it is created from the compiled-in defaults. If a file exists but does not parse, the function returns an error, and in that case the settings are not silently replaced with the defaults.

Saving settings also has a separate method for each section:

`SaveGeneral`, `SaveWriter`, `SaveICMP`, `SaveTCP`, `SaveHTTP`, `SaveXray`, and `SaveDNS`.

The inspector for each section changes the settings struct in memory and calls the matching method when saving.

The interface provides the settings and the Store to different parts through `ui.Session`:

```go
type Session struct {
    Config  *config.ScannerConfig
    Store   *config.Store
    Program ui.Program
    // env and core logger
}
```

The `Session.Env()` method returns interface environment details, including the layout, theme, keys, interface logger, and `UISettings`.

### Settings validation

The `config/validate` package has a separate file for each settings section.

- The `ValidateXxx` functions return a map of field errors without changing any values.
- The `NormalizeXxx` functions reset invalid values to their defaults and produce a `Warning` for each correction.
- The `aggregate.go` file combines these functions in `ValidateAll` and `NormalizeAll`.

At startup, `NormalizeAll` runs and shows the corrections to the user. Corrected values stay in memory and are written to disk only when the user saves that section through the inspector.

## IP lists

- `parser.go` has the `StreamActiveIPs(ctx, path, limit, shuffled, out)` function, which the Engine uses to send IPs to workers. The `StreamCIDR` function expands a network prefix.
- `csv.go` and `parser.go` read the two column CSV format `<ip_or_cidr>,<enabled>`.
- `loader.go` loads external lists and provides the `CountIPs` and `CountActiveIPs` functions to calculate the total number of IPs.
- `registry.go` lists the files under `ips/` and `shuffle.go` randomizes the input order.

Disabled entries are skipped when streaming. Both IPv4 and IPv6 prefixes are supported, so counting the IPs in a large IPv6 prefix may reach a saturated value instead of an exact count.

## DNS subsystem

- `query.go` builds and sends DNS queries over UDP, TCP, and DoT.
- `types.go` parses transports, record types, rcode values, authentication methods, and tunnel protocol names.
- The files `dnstt.go`, `vaydns.go`, `slipstream.go`, `masterdns.go`, `stormdns.go`, and `thefeed.go` define the config structs and service interfaces for each protocol and manage the settings files under `assets/dns-tunneling/`. The interfaces are `DNSTTService`, `VayDNSService`, `SlipstreamService`, `MasterDNSService`, `StormDNSService`, and `TheFeedService`.
- `store.go` provides the general config file store.
- `registry.go` has the `GetAllDNSTunsFile` aggregator, which collects the settings for all six services.
- `validate.go` holds the shared validation helpers.
- `embedded.go` wraps the internal MasterDNS and StormDNS clients.
- The `socks/` package provides a SOCKS5 client to validate a tunnel after the connection is up.
- The `ssh/` package provides an SSH client for tunneled connections through an SSH proxy.

## Xray integration

- `service.go` runs and manages Xray instances in the same process through the internal `xray-core` library. The `inbound.go` and `outbound.go` files build the inbound and outbound JSON settings.
- `link.go` parses share links.
- The `download.go`, `latency.go`, and `upload.go` files in the `speedtest` package measure connection speed and latency.

For the selected outbound, `xrayprobe` builds the needed settings, leases a port, starts an Xray instance in the same process, and measures latency through the local proxy. It runs a speed test when needed and releases the resources it used at the end.

## Process management

The `internal/core/process` package holds the old cross platform process tracking tool, including `process_unix.go` and `process_windows.go`.

The current scanner does not use this package, but its code is still in the project. The only child processes the scanner runs are the `termux-wake-lock` and `termux-wake-unlock` commands on Android.

## Logger

| Logger | File | Purpose |
|---|---|---|
| Core | `logs/core.log` | Engine, probes, settings, and result input and output operations |
| UI | `logs/ui.log` | Component lifecycle, file operations, and interface errors |

Both loggers read `log_level` from `settings/ui_settings.toml`. Besides writing to files, they also send logs to subscribers that receive them live.

Log file rotation uses lumberjack. The current settings are a 50 megabyte size cap, three backup versions kept, a maximum of 7 days of retention, and compression of old files.

## Startup

The Startup stage is implemented inside the TUI at `internal/ui/main/startup`. It runs after the Splash animation and before the Workspace appears, and it shows the status of each check live in the status bar.

The checks are defined as private functions in `checklist.go`. The Startup component manages the process with its `Init` and `Update` methods and runs the checks in order inside a Goroutine.

The order of the checks is:

```text
1. Logger       Check the Core and UI loggers and run core.Init()
                to register the probe result schemas

2. Config       Run store.Load() and report an error if the TOML
                is invalid, then run validate.NormalizeAll to
                correct out of range values

3. Xray         Show the internal core version and check the outbounds

4. DNSTT        Check the DNSTT settings files

5. Slipstream   Find and load the client library
                and check the settings files

6. VayDNS       Check the VayDNS settings files

7. MasterDNS    Check the MasterDNS settings files

8. StormDNS     Check the StormDNS settings files

9. TheFeed      Check the TheFeed settings files

10. App         Wait for the user to press Enter
```

Each check reports its status through a `reporter`. Messages can have one of these levels: `[INFO]`, `[SUCCESS]`, `[WARN]`, or `[ERROR]`.

Critical errors stop the following checks from running. A missing Slipstream library reports an error but does not stop Startup, and in that case only the Slipstream scan cannot run.

When the checks finish, the user enters the Workspace by pressing Enter.

In the current structure, `main.go` does not load settings and does not run the Startup checks. Its job is to build the loggers, create the application (which sets up the theme and interface settings), and start the Bubble Tea program.

Errors from reading settings, registering schemas, loading the library, and validating config appear in the Startup interface itself, not as silent errors before the TUI comes up. Corrected values also stay in memory so the user can save that section through the inspector if needed.

## Related pages

- [Architecture](./architecture/): overall project structure and layers
- [User interface](./ui/): component architecture and themes
