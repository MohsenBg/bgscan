---
title: "General Settings"
weight: 3
---

# General Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **General Settings** in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/general_settings.toml`

Global scan limits, probe rate limiting, pipeline execution mode, and buffer sizing.

## Quick Reference

| Setting | Default | Description |
| --------- | --------- | ------------- |
| `status_interval` | `1000` | Progress update interval in milliseconds |
| `stop_after_found` | `0` | Reserved; not enforced by the current engine |
| `max_ips_to_test` | `0` | Cap on IPs read from the input source |
| `pipeline_mode` | `"streaming"` | Execution mode for multi-stage scans |
| `max_ips_per_stage` | platform-dependent | Channel buffer size between streaming stages |
| `batch_size` | platform-dependent | Batch size for batch mode |
| `shuffled` | `true` | Randomize target order |
| `min_probe_duration` | platform-dependent | Minimum time spent per probe (ms) |
| `probe_per_sec` | platform-dependent | Sustained probe rate ceiling (token-bucket refill) |
| `probe_burst` | platform-dependent | Max probes in a burst (token-bucket capacity) |

## Status Interval

```toml
status_interval = 1000
```

How often each stage emits a progress snapshot to the UI, in milliseconds. Valid range is 100 ms to 60 s. Lower values give smoother feedback and cost more CPU on very large scans.

## Stop After Found

```toml
stop_after_found = 0
```

Intended to halt a scan after N successful results. The field is loaded, validated, and editable in the inspector, but the scan engine does not act on it yet. Leave it at `0`.

## Max IPs to Test

```toml
max_ips_to_test = 0
```

Caps how many IPs are read from the input source. `0` reads everything. Useful when sampling a large list or a wide IPv6 prefix, where scanning to completion is not practical.

## Pipeline Mode

```toml
pipeline_mode = "streaming"
```

Controls how targets flow through a multi-stage scan.

| Value | Behavior |
| --- | --- |
| `streaming` (alias `parallel`) | All stages run concurrently. Successful IPs move to the next stage through in-memory channels. |
| `sequential` (alias `simple`) | Stage N finishes and writes its result file. Stage N+1 reads that file as input. |
| `batch` (alias `pipeline`) | IPs are chunked. Each chunk traverses every stage before the next chunk is read. |

Unrecognized or empty values fall back to `sequential` at parse time, but the config validator rejects anything outside the list above and restores the default first. See [Scan Pipeline](../scanner/scan-pipeline.md) for the trade-offs.

## Max IPs Per Stage

```toml
max_ips_per_stage = 100000
```

Buffer size of the channel connecting one streaming stage to the next. A larger buffer means a fast stage blocks less often when the following stage is slower, at the cost of memory. Each buffered entry is a single address.

The value is raised to the next stage's worker count when that count is higher. Only used in `streaming` mode. Valid range is 1 to 10,000,000.

## Batch Size

```toml
batch_size = 5000
```

Number of IPs per chunk in `batch` mode. With more than one stage, the effective size is `max(batch_size, highest worker count among stages after the first)`, so a batch is never smaller than the workers that have to drain it.

## Shuffle Targets

```toml
shuffled = false
```

Randomizes target order before scanning. Useful to spread load across subnets, avoid hammering one range, and get a representative sample when combined with `max_ips_to_test`.

## Probe Rate Limit

```toml
probe_per_sec = 1500
probe_burst = 200
```

`probe_per_sec` and `probe_burst` form a token-bucket limiter applied to every probe across all stages. `probe_per_sec` is the sustained ceiling — tokens refilled per second. `probe_burst` is the bucket capacity — how many probes may fire back-to-back before the limiter blocks. Together they cap outgoing request volume regardless of `Workers`, which is how you avoid tripping upstream rate limits or DPI.

Both are validated. `probe_per_sec` ranges 1 to 1,000,000; `probe_burst` ranges 1 to 10,000. Out-of-range values are rejected, the default is restored, and the correction is logged. The effective default is platform-dependent (see the Quick Reference table).

## Minimum Probe Duration

```toml
min_probe_duration = 10
```

Enforces a floor on how long each probe takes, in milliseconds. After a probe finishes faster than this, the worker sleeps for the remainder before moving to the next target. Used to normalize timing-based side channels and to pace scans that would otherwise complete too fast.

Valid range is 10 ms to 5 s. The effective default is platform-dependent (see the Quick Reference table).

## Related Files

- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS/HTTP3 probe configuration
- [`dns_settings.toml`](./dns.md) — DNS scan configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
- [`writer_settings.toml`](./writer.md) — Result output configuration
