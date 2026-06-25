<div align="left">

[**English**](./writer_settings.md)  |  [**فارسی**](./writer_settings.fa.md)

</div>

# Result Writer Configuration

Configuration file: `settings/result_writer.toml`

This file controls how scan results are collected from worker goroutines, accumulated in memory, and flushed to disk.

---

## Settings

### `merge_flush_interval`

```toml
merge_flush_interval = 2000
```

Interval in **milliseconds** at which delta results are merged into the main result file.

| Value  | Behavior                                         |
| ------ | ------------------------------------------------ |
| `1000` | Merge every 1 second — lower latency, more I/O   |
| `2000` | Merge every 2 seconds (default)                  |
| `5000` | Merge every 5 seconds — less I/O, higher latency |

---

### `chan_size`

```toml
chan_size = 4096
```

Capacity of the internal channel used by scanner workers to send `IPScanResult` entries to the writer goroutine.

A larger buffer reduces the chance of workers blocking while the writer is busy flushing to disk.

| Value              | Trade-off                                      |
| ------------------ | ---------------------------------------------- |
| Low (`512`–`1024`) | Less memory, workers may block under high load |
| Medium (`4096`)    | Balanced default                               |
| High (`8192`+)     | Fewer stalls, higher memory usage              |

---

### `batch_size`

```toml
batch_size = 4096
```

Initial capacity of the in-memory batch used to accumulate `IPScanResult` entries before flushing to disk.

This is a pre-allocation hint — the batch can grow beyond this value if needed, but starting at the right size avoids repeated memory reallocations during high-throughput scans.

| Value              | Trade-off                                     |
| ------------------ | --------------------------------------------- |
| Low (`512`–`1024`) | Lower memory footprint, more frequent flushes |
| Medium (`4096`)    | Balanced default                              |
| High (`8192`+)     | Fewer flushes, higher peak memory usage       |

---

## Full Example

```toml
# ── Result Writer ─────────────────────────────────────────────
merge_flush_interval = 2000   # Merge delta results every 2 seconds
chan_size             = 4096   # Channel buffer between workers and writer
batch_size           = 4096   # In-memory batch capacity before disk flush
```
