---
title: "Result Writer Settings"
weight: 4
---

# Result Writer Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **General Settings**  tab 2 in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/writer_settings.toml`

This file controls buffering and how often results are flushed to disk.

## Quick Reference

| Setting | Default | Description |
|---------|---------|-------------|
| `merge_flush_interval` | `2000` | Interval (in milliseconds) for merging delta results into the main result file. |
| `chan_size` | `4096` | Capacity of the internal channel used by scanner workers to send IP scan results to the writer goroutine. |
| `batch_size` | `4096` | Initial capacity of the in-memory batch used to accumulate IP scan results before flushing them to disk. |

---

## Merge Flush Interval

Controls how frequently (in milliseconds) partial scan results are committed to disk. Lower values increase disk I/O but reduce memory usage. Higher values reduce disk I/O but increase memory usage.

```toml
merge_flush_interval = 2000
```

**Recommended values:**

- `1000-5000` for typical use
- Lower values for systems with fast storage
- Higher values for systems with limited I/O bandwidth

---

## Channel Size

The capacity of the internal channel used by scanner workers to send IP scan results to the writer goroutine.
Higher values can help prevent worker bottlenecks during heavy disk I/O.

```toml
chan_size = 4096
```

**Recommended values:**

- `1024-4096` for moderate throughput
- `4096-16384` for high-throughput scanning

---

## Batch Size

The initial capacity of the in-memory batch used to accumulate IP scan results before flushing them to disk.
Higher values reduce the frequency of disk writes but increase memory usage.

```toml
batch_size = 4096
```

**Recommended values:**

- `1024-4096` for memory-constrained systems
- `4096-16384` for systems with ample memory

---

## Related Files

- [`general_settings.toml`](./general.md) — Global scan control and pipeline mode
- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS/HTTP3 probe configuration
- [`dns_settings.toml`](./dns.md) — DNS scan configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation

