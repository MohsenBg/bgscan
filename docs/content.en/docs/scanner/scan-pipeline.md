---
title: "Scan Pipeline"
weight: 3
---

# Scan Pipeline

A pipeline chains scan stages so that only the targets surviving one stage reach the next. Stages come from the scan type you pick: a DNS tunneling scan with resolver pre-scan builds two, an Xray scan with a pre-scan builds two, and a plain ICMP scan builds one.

A single-stage scan skips the chain logic entirely. The pipeline mode only matters once there are two or more stages.

## Pipeline Modes

Set in [`general_settings.toml`](../settings/general.md):

```toml
pipeline_mode = "streaming"
```

### Streaming

```
[ICMP] ──chan──> [TCP] ──chan──> [HTTP]
   │               │               │
   ▼               ▼               ▼
result/icmp/   result/tcp/    result/http/
```

Every stage runs at the same time with its own worker pool. A target that passes a stage is pushed straight into the next stage's channel, so later stages start working before earlier ones finish.

Channel capacity comes from `max_ips_per_stage`, raised to the next stage's worker count when that is higher. When a downstream stage falls behind, its channel fills and the upstream stage blocks, which bounds memory.

Fastest option, and the default. Memory scales with the channel buffers.

### Sequential

```
[ICMP] ──> result/icmp/ ──> [TCP] ──> result/tcp/ ──> [HTTP] ──> result/http/
```

Each stage runs to completion and writes its result file. The next stage reads that file as its input. Nothing runs concurrently across stages.

Lowest memory, slowest wall-clock time, since total time is the sum of all stages. If a stage produces no results, the chain stops there.

Note that `sequential` is what an unrecognized mode string falls back to at parse time.

### Batch

```
batch 1 ──> [ICMP] ──> [TCP] ──> [HTTP]
batch 2 ──> [ICMP] ──> [TCP] ──> [HTTP]
...
```

Targets are read in chunks of `batch_size`. A chunk passes through every stage before the next chunk is read. Survivors are handed between stages in memory, without a round trip through disk.

With more than one stage the effective chunk size is `max(batch_size, highest worker count among the stages after the first)`, so a chunk is never too small to keep the later stages busy. Memory sits between the other two modes and stays predictable regardless of input size.

## Data Flow

Only targets a stage accepts are forwarded. What counts as accepted is the probe's own success condition:

- ICMP: an echo reply arrived within the timeout
- TCP: the handshake completed on the configured port
- HTTP: a response arrived and its status code is in `accepted_status_codes`
- DNS resolver: the response code is in `accepted_rcodes`, and the DPI check passed when enabled
- DNSTT, VayDNS, and Slipstream: the tunnel came up and validated through the local SOCKS5 port
- Xray: the proxy connected, and any enabled speed test met its minimum

Each stage writes its own result file regardless of what happens downstream, so intermediate output is always available for a later re-scan.

## Example

Starting from 10,000 IPs with an ICMP → TCP → HTTP chain:

| Stage | Input | Passes | Result file |
|---|---|---|---|
| ICMP | 10,000 | 2,000 | `result/icmp/` |
| TCP | 2,000 | 500 | `result/tcp/` |
| HTTP | 500 | 300 | `result/http/` |

Each result file contains that stage's passing targets, not everything it examined.

## Stage Configuration

Every stage reads its own settings file. Worker counts, timeouts, and retries are per stage, so a cheap first stage can run far wider than an expensive last one.

- [`icmp_settings.toml`](../settings/icmp.md)
- [`tcp_settings.toml`](../settings/tcp.md)
- [`http_settings.toml`](../settings/http.md)
- [`dns_settings.toml`](../settings/dns.md)
- [`xray_settings.toml`](../settings/xray.md)

## Related Topics

- [Scan Types](./scan-types.md)
- [Result Files](./result-files.md)
- [General Settings](../settings/general.md)
