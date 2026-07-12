---

title: "General Settings"
weight: 3
---

# General Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **General Settings** in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/general_settings.toml`

This file controls the core behavior of the bgscan scanner, including scan limits, execution modes, and performance tuning.

## Quick Reference

| Setting | Default | Description |
|---------|---------|-------------|
| `status_interval` | `1000` | UI refresh rate in milliseconds |
| `stop_after_found` | `0` | Stop after finding N successful targets |
| `max_ips_to_test` | `0` | Maximum number of IPs to scan |
| `shuffled` | `true` | Randomize target order |
| `pipeline_mode` | `"streaming"` | Execution mode for multi-stage scans |
| `max_ips_per_stage` | `100000` | Memory cap per pipeline stage |
| `batch_size` | `1000` | Batch size for batch mode |

---

## Scan Control

### Status Interval

```toml
status_interval = 1000
```

Controls how often (in milliseconds) the UI refreshes with progress updates. Lower values give smoother real-time feedback but use more CPU. Higher values reduce overhead for very large scans.

**Recommended values:**

- `500-1000` — Smooth updates for most scans
- `2000-5000` — Large scans where performance matters more than live updates

---

### Stop After Found

```toml
stop_after_found = 0
```

Stops the scan automatically after finding a specified number of successful targets. A "successful target" is one that passes all scan conditions (e.g., responds to ICMP, has an open port).

**Behavior:**

- `0` — Disabled (default). Scans all targets in the list.
- `N` — Stops immediately after finding N successful results.

**Use cases:**

- Finding the first few working proxies quickly
- Limiting scan time when you only need a handful of results
- Testing if any targets are alive before committing to a full scan

---

### Max IPs to Test

```toml
max_ips_to_test = 0
```

Limits the total number of IPs read from your input source (file or manual entry). Useful for sampling large IP lists or testing with a subset before running a full scan.

**Behavior:**

- `0` — No limit (default). Reads all available IPs.
- `N` — Reads and scans at most N IPs.

**Use cases:**

- Testing configuration on a small sample (e.g., first 100 IPs)
- Reducing scan time on massive IP lists
- Debugging without waiting for thousands of targets

---

### Shuffle Targets

```toml
shuffled = true
```

Randomizes the order of target IPs before scanning begins. This helps:

- Avoid overwhelming specific subnets
- Reduce the chance of triggering firewall rate-limits or IDS alerts
- Get a more representative sample when using `max_ips_to_test`

**Recommendation:** Keep enabled (`true`) for most scans. Disable only if you need deterministic, repeatable scan order.

---

## Pipeline Execution Mode

bgscan supports multi-stage scanning (e.g., ICMP → TCP → HTTP). The pipeline mode controls how targets flow through these stages.

### Pipeline Mode

```toml
pipeline_mode = "streaming"
```

**Available modes:**

| Mode | Description | RAM Usage | Speed | Best For |
|------|-------------|-----------|-------|----------|
| `"streaming"` | All stages run simultaneously, streaming IPs via channels | High | Fastest | Large datasets, maximum throughput |
| `"sequential"` | Each stage waits for the previous to finish; uses disk | Low | Slowest | Memory-constrained systems |
| `"batch"` | IPs processed in batches through all stages | Medium | Balanced | Predictable memory usage |

**Default:** `"streaming"` — Recommended for most users with adequate RAM.

---

### Max IPs Per Stage

```toml
max_ips_per_stage = 100000
```

Sets a hard memory cap on how many IPs each pipeline stage can hold at once. When this limit is reached, the previous stage pauses until space becomes available.

**Purpose:** Prevents out-of-memory crashes on very large scans.

**Tuning:**

- Reduce to `50000` or lower on systems with <4GB RAM
- Increase for systems with abundant memory and very large IP lists

---

### Batch Size

```toml
batch_size = 1000
```

Defines how many IPs are processed together when using `"batch"` pipeline mode. Each batch goes through all scan stages before the next batch begins.

**Tuning:**

- Smaller batches (`500-1000`) — Lower memory, more frequent disk I/O
- Larger batches (`5000-10000`) — Better throughput, higher memory usage

**Note:** Only applies when `pipeline_mode = "batch"`.

---

## Related Files

- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS scan configuration
- [`dns_settings.toml`](./dns.md) — DNS scan configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
- [`writer_settings.toml`](./writer.md) — Result output configuration
