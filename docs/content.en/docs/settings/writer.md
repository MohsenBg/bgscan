---
title: "Result Writer Settings"
weight: 4
---

# Result Writer Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **General Settings** tab 2 in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/writer_settings.toml`

Buffering and flush behavior of the result writer, plus where result files are stored.

Each scan stage owns one writer. Workers push results into a channel, the writer accumulates them into a batch, and each flush merge-sorts that batch into the stage's CSV file.

## Quick Reference

| Setting | Default | Description |
| --------- | --------- | ------------- |
| `merge_flush_interval` | `2000` | Flush interval in milliseconds |
| `chan_size` | platform-dependent | Capacity of the worker to writer channel |
| `batch_size` | platform-dependent | Results buffered in memory before a flush |
| `result_directory` | `"result"` | Base directory for result files |

## Merge Flush Interval

```toml
merge_flush_interval = 2000
```

Time between periodic flushes, in milliseconds. Valid range is 100 ms to 5 minutes. A flush also happens as soon as `batch_size` results accumulate, and once more when the writer stops, so nothing accepted before shutdown is lost.

Shorter intervals write results to disk sooner and rewrite the file more often. Longer intervals cut disk I/O and hold more results in memory.

## Channel Size

```toml
chan_size = 1024
```

Capacity of the channel workers use to hand results to the writer goroutine. When the channel fills, workers block until the writer catches up. Raise it if a stage produces results in bursts faster than the merge can absorb. Valid range is 1 to 1,000,000.

## Batch Size

```toml
batch_size = 4096
```

Number of results held in memory before a flush is forced. It also sets the initial batch slice capacity and the size of the buffered writer used during the merge. Valid range is 1 to 1,000,000.

## Result Directory

```toml
result_directory = "result"
```

Base directory, relative to the bgscan binary, holding one subdirectory per result schema: `result/icmp/`, `result/tcp/`, `result/http/`, `result/xray/`, `result/dns_resolver/`, `result/dnstt/`, and `result/slipstream/`. Directories are created on demand.

The value must be a plain directory name, not a path.

## Related Files

- [`general_settings.toml`](./general.md) — Global scan control and pipeline mode
- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS/HTTP3 probe configuration
- [`dns_settings.toml`](./dns.md) — DNS scan configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
