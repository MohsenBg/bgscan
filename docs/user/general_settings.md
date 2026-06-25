<div align="right">
  
  [**English**](./general_settings.md) &nbsp;|&nbsp; [**فارسی**](./general_settings.fa.md)
  
</div>

# General Settings

Configuration file: `settings/general_settings.toml`

This file controls the scanner's core behavior — from scan limits and IP ordering to pipeline execution strategy and memory management.

---

## Table of Contents

- [Scan Control](#scan-control)
- [Pipeline Execution](#pipeline-execution)
- [Pipeline Tuning](#pipeline-tuning)

---

## Scan Control

These settings define how many IPs are scanned, in what order, and when to stop.

### `status_interval`

```toml
status_interval = 1000
```

Interval in **milliseconds** between status updates pushed to the UI.

| Value | Behavior |
|-------|----------|
| `1000` | Update every 1 second (default) |
| `500` | Update twice per second |

---

### `stop_after_found`

```toml
stop_after_found = 0
```

Stops scanning after this many **successful hits** (IPs that pass all scan conditions).

| Value | Behavior |
|-------|----------|
| `0` | Disabled — scan all IPs |
| `N` | Stop after N successful results |

> **Note:** In chain-mode scanning, this limit applies **only to the final stage**. Multi-stage support is not yet implemented.

---

### `max_ips_to_test`

```toml
max_ips_to_test = 0
```

Caps the number of IPs read from the input source. Useful for sampling large IP lists or limiting resource-heavy scans.

| Value | Behavior |
|-------|----------|
| `0` | No limit — read all IPs |
| `N` | Read at most N IPs |

---

### `shuffled`

```toml
shuffled = true
```

Randomizes the target IP order before scanning begins.

**Why enable this?**
- Prevents consecutive hits to the same subnet or network block
- Reduces the chance of triggering rate-limiting or firewall alerts
- Distributes scan load more evenly across the network

---

## Pipeline Execution

### `pipeline_mode`

```toml
pipeline_mode = "streaming"
```

Defines how multi-stage scanning is executed. Each mode offers a different tradeoff between throughput, memory usage, and predictability.

| Mode | Description | Best For |
|------|-------------|----------|
| `"sequential"` | Stages run one at a time; results written to disk between stages | Low-RAM environments, small scans |
| `"streaming"` | All stages run in parallel; IPs flow through buffered channels | Maximum throughput, large datasets |
| `"batch"` | IPs are grouped into batches; each batch passes through all stages | Uneven stage costs, predictable memory |

**Mode details:**

- **`sequential`** — Safest memory profile. Each stage waits for the previous to fully complete before starting. Slowest due to disk I/O overhead.

- **`streaming`** *(default)* — Highest performance. Stages are connected by in-memory buffered channels and process IPs concurrently. Ideal for real-time, high-volume scanning.

- **`batch`** — Hybrid approach. IPs are chunked into batches (see [`batch_size`](#batch_size)); each batch flows through all stages before the next batch begins. More memory-predictable than `streaming`.

---

## Pipeline Tuning

### `max_ips_per_stage`

```toml
max_ips_per_stage = 100000
```

The maximum number of IPs a single stage can hold at one time in chain mode.

When this limit is reached, the upstream stage **blocks** until space becomes available — acting as a hard backpressure cap. This prevents unbounded memory growth in long-running scans.

> This is a fixed hard cap, not a soft hint.

---

### `batch_size`

```toml
batch_size = 1000
```

Number of IPs per batch when `pipeline_mode = "batch"`.

Each group of `batch_size` IPs passes through all pipeline stages sequentially before the next group begins.

> **Only applies when `pipeline_mode = "batch"`.**

---

## Full Example

```toml
# ── Scan Control ──────────────────────────────────────────────────────
status_interval   = 1000   # Push UI updates every 1 second
stop_after_found  = 100    # Stop after 100 successful hits
max_ips_to_test   = 50000  # Cap input at 50k IPs
shuffled          = true   # Randomize target order

# ── Pipeline ──────────────────────────────────────────────────────────
pipeline_mode     = "streaming"
max_ips_per_stage = 100000
batch_size        = 1000
```
