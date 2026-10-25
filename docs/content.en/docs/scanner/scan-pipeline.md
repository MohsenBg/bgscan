---
title: "Scan Pipeline"
weight: 3
---

# Scan stages

BGScan can connect several scan stages so that only the targets accepted by one stage reach the next. The scan engine supports multi stage chains, but in version 2.12 the terminal interface (TUI) builds the chain based on the chosen scan type:

- In a DNS tunnel scan, if `check_dns_resolver = true`, a resolver check stage runs before the tunnel starts.
- In an Xray scan, if `pre_scan_type` is `tcp`, `icmp`, or `http`, that scan runs before Xray.
- The other scans have only one stage.

Even single stage scans use this same mechanism. The difference between execution modes appears when a scan has several stages and targets move between them.

## Scan execution modes

You set the execution mode in the [`general_settings.toml`](../settings/general.md) file:

```toml
pipeline_mode = "streaming"
```

### Streaming

```text
[TCP pre-scan] ──chan──> [Xray]
      │                    │
      ▼                    ▼
result/tcp/          result/xray/
```

All stages run at the same time and each has its own worker pool. Any target accepted by a stage is sent straight to the next stage through a channel. As a result, the next stage does not need to wait for the earlier stage to finish.

The channel capacity comes from `max_ips_per_stage`. If the next stage has more workers, the channel capacity grows to that count. When the next stage cannot process targets fast enough, its channel fills and the earlier stage pauses temporarily. This helps keep memory use under control.

This mode is the fastest option and the default. Memory use depends on the size of the channel buffers.

### Sequential

```text
[TCP pre-scan] ──> result/tcp/ ──> [Xray] ──> result/xray/
```

Each stage runs to completion and saves its results in its own file. The next stage reads that same file as input. The stages therefore do not run at the same time.

This mode uses the least memory, but it is usually slower, because the stages run one after another and their run times add up. If a stage produces no results, the remaining stages do not run.

**Note about the mode names:** If the `pipeline_mode` value is unknown when parsed, it falls back to `sequential`. The settings validator also accepts the values `simple`, `parallel`, and `pipeline`, but the engine does not convert these names to the other modes and they run as `sequential` too.

To set this option, use the exact names `streaming`, `sequential`, or `batch`. Any other invalid value falls back to the default `streaming` mode during the initial program check, before the scan starts.

### Batch

```text
batch 1 ──> [TCP pre-scan] ──> [Xray]
batch 2 ──> [TCP pre-scan] ──> [Xray]
...
```

In this mode the inputs are split into parts with the size `batch_size`. Each part passes through all the stages and then the next part takes its turn. Targets accepted by a stage are kept in memory until they reach the next stage, so they do not need to be written to disk and read back to move between stages.

When there is more than one stage, the actual size of each part is the larger of `batch_size` and the highest worker count among the stages after the first stage. This way the parts do not get so small that the next stage sits idle for lack of input.

Memory use in this mode usually falls between `streaming` and `sequential` and is more predictable with the part size, even when the number of inputs is large.

## Passing targets between stages

Only targets accepted by a stage reach the next one. The acceptance rule depends on the probe type:

- **ICMP:** An echo reply arrives before the timeout.
- **TCP:** The connection and handshake complete on the configured port.
- **HTTP:** A response arrives and its status code is in `accepted_status_codes`.
- **DNS resolver:** The response code is in `accepted_rcodes` and, when the DPI check is enabled, that check also succeeds.
- **DNS tunnel:** The DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, or TheFeed tunnel comes up and validates. Slipstream, MasterDNS, and StormDNS check the connection through a local SOCKS5 listener.
- **Xray:** The proxy connection is made and any enabled speed test meets the configured minimum speed.

Each stage saves its own result file, even if later stages fail or produce no results. You can therefore use the results of each stage separately for later scans.

## Example

Suppose we start a scan with 10,000 IPs and have a TCP pre-scan before Xray:

| Stage | Input | Accepted | Result file |
|---|---:|---:|---|
| TCP pre-scan | 10,000 | 2,000 | `result/tcp/` |
| Xray | 2,000 | 500 | `result/xray/` |

Each stage result file keeps only the targets that stage accepted, not every target it checked.

## Settings for each stage

Each stage has its own settings. You can set the worker count, timeout, and attempt count separately for each stage. This way a cheap first stage can check a large number of IPs and the final stage runs on a smaller set of targets.

- [`icmp_settings.toml`](../settings/icmp.md)
- [`tcp_settings.toml`](../settings/tcp.md)
- [`http_settings.toml`](../settings/http.md)
- [`dns_settings.toml`](../settings/dns.md)
- [`xray_settings.toml`](../settings/xray.md)

## Related topics

- [Scan types](./scan-types.md)
- [Result files](./result-files.md)
- [General settings](../settings/general.md)
