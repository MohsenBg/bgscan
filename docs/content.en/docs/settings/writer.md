---
title: "Result Writer Settings"
weight: 4
---

# Result writer settings

> 💡 **Note:** You do not need to edit the TOML file by hand. You can change these settings directly inside BGScan.
>
> Go to **Settings → General Settings** from the main menu, open the **Writer** tab, and change the options with the Inspector inside the TUI.

Settings file: `settings/writer_settings.toml`

This section defines the result buffering settings, the flush timing, and the directory where result files are saved.

Each scan stage has its own Writer. The workers send results to the Writer through a channel. The Writer collects the results in a batch and on each flush merge-sorts them into the CSV file for that stage.

## Quick reference

| Setting | Default | Description |
|---|---|---|
| `merge_flush_interval` | `2000` | The time between flushes, in milliseconds |
| `chan_size` | Depends on the platform | The capacity of the channel between the workers and the Writer |
| `batch_size` | Depends on the platform | The number of results collected in memory before a flush |
| `result_directory` | `"result"` | The main directory for result files |

## Time between flushes

```toml id="2f7x9c"
merge_flush_interval = 2000
```

Sets the time between periodic flushes in milliseconds. The allowed range is from 100 milliseconds to 5 minutes.

Besides the periodic flush, another flush happens when the number of results reaches `batch_size`. A final flush also happens when the Writer stops so results received up to that point are not lost.

- **Shorter interval:** Results are written to disk sooner, but the CSV file is rewritten and sorted more often.
- **Longer interval:** There are fewer disk operations, but more results stay in memory for longer.

## Channel size

```toml id="9g5v2k"
chan_size = 1024
```

Sets the capacity of the channel the workers use to send results to the Writer.

When the channel fills, the workers wait until there is free capacity. If results are produced faster than the Writer can handle them, you can raise this value.

The allowed range is `1` to `1000000`.

## Batch size

```toml id="4n8w6j"
batch_size = 4096
```

Sets the number of results collected in memory so that a flush happens once this count is reached.

This value also affects the initial capacity of the batch slice and how much memory is used to hold results during the merge.

The allowed range is `1` to `102400`.

## Result directory

```toml id="7m3q5v"
result_directory = "result"
```

Sets the main directory for result files. The path of this directory is relative to the BGScan binary and a separate subdirectory is created for each result type:

- `result/icmp/`
- `result/tcp/`
- `result/http/`
- `result/xray/`
- `result/dns_resolver/`
- `result/dnstt/`
- `result/vaydns/`
- `result/slipstream/`
- `result/masterdns/`
- `result/stormdns/`
- `result/thefeed/`

These directories are created automatically when needed.

**Note:** The `result_directory` value must be only a directory name, not a full path.

## Related files

- [`general_settings.toml`](./general.md): general scan settings and the scan execution mode
- [`icmp_settings.toml`](./icmp.md): ICMP scan settings
- [`tcp_settings.toml`](./tcp.md): TCP port scan settings
- [`http_settings.toml`](./http.md): HTTP, HTTPS, and HTTP/3 probe settings
- [`dns_settings.toml`](./dns.md): DNS scan settings
- [`xray_settings.toml`](./xray.md): Xray scan settings and outbound validation
