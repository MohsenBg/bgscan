---
title: "General Settings"
weight: 3
---

# General settings

> 💡 **Note:** You do not need to edit the TOML file by hand. You can change these settings directly inside BGScan.
>
> Go to **Settings → General Settings** from the main menu and set the options with the Inspector inside the TUI.

Settings file: `settings/general_settings.toml`

This file defines the general scan settings, including the limit on the number of IPs, the probe rate, how multi stage scans run, and the buffer sizes.

## Quick reference

| Setting | Default | Description |
|---|---|---|
| `status_interval` | `1000` | How often the scan status updates, in milliseconds |
| `max_successful_ips` | `0` | Stops the scan after reaching a set number of successful results. `0` means no limit |
| `max_ips_to_test` | `0` | The maximum number of IPs read from the scan input |
| `pipeline_mode` | `"streaming"` | How the multi stage scan runs |
| `max_ips_per_stage` | Depends on the platform | The channel buffer size between stages in Streaming mode |
| `batch_size` | Depends on the platform | The size of each part in Batch mode |
| `shuffled` | `true` | Shuffles the target order randomly |
| `min_probe_duration` | Depends on the platform | The minimum time each probe runs, in milliseconds |
| `probe_per_sec` | Depends on the platform | The allowed sustained probe rate |
| `probe_burst` | Depends on the platform | The maximum number of probes back to back |

## Status update interval

```toml
status_interval = 1000
```

This option defines how often each stage sends its progress status to the interface, in milliseconds. The allowed range is from 100 milliseconds to 60 seconds.

A lower value makes the scan status update more smoothly, but on very large scans it may increase CPU use.

## Successful IP limit

```toml
max_successful_ips = 0
```

The number of successful results after which the scan stops.

For example, if you set the value to `1000`, the scan stops after reaching 1000 successful results. A value of `0` means the scan continues until every target is checked.

## Target IP limit

```toml
max_ips_to_test = 0
```

The maximum number of IPs read from the scan input.

A value of `0` means there is no limit. This option is useful for sampling large lists or wide IPv6 prefixes, especially when checking every address would take a long time.

## Scan execution mode

```toml
pipeline_mode = "streaming"
```

This option defines how targets move between stages in a multi stage scan.

| Value | Behavior |
|---|---|
| `streaming` | All stages run at the same time and successful IPs move to the next stage through in-memory channels. |
| `sequential` | Each stage runs to completion and saves its results in its own file. The next stage reads the result file of the earlier stage as input. |
| `batch` | IPs are split into smaller parts. Each part passes through all the stages and then the next part is processed. |

Use these three exact values: `streaming`, `sequential`, and `batch`.

The settings validator also accepts the values `simple`, `parallel`, and `pipeline`, but the scan engine does not convert them to the other modes, so they run as `sequential`.

Invalid values also fall back to the default `streaming` mode during the initial settings check before the scan starts, and the change is reported.

To learn more about the differences between these modes, see the [Scan stages](../scanner/scan-pipeline.md) page.

## IP limit per stage

```toml
max_ips_per_stage = 100000
```

This option defines the buffer size of the channel that carries the output of one stage to the next in `streaming` mode. Each buffer entry holds one IP address.

A larger buffer means a faster stage waits less often for a slower stage, but it also increases memory use.

If the next stage has more workers than the set value, the buffer capacity grows to that stage's worker count.

This option is used only in `streaming` mode. The allowed range is from 1 to 10,000,000.

## Part size in Batch mode

```toml
batch_size = 5000
```

The number of IPs processed in each part of `batch` mode.

If the scan has more than one stage, the actual size of each part is the largest value between `batch_size` and the highest worker count among the stages after the first stage:

`max(batch_size, highest worker count among stages after the first)`

This way the part size is never smaller than the number of workers in the later stages that need input.

## Shuffling the target order

```toml
shuffled = true
```

When this option is enabled, the target order is changed randomly before the scan.

This keeps requests from piling up on one range and spreads the scan load across different subnets. If you use this option with `max_ips_to_test`, you can also check a sample from a large list of targets.

## Probe rate limit

```toml
probe_per_sec = 500
probe_burst = 100
```

These two options together build a Token Bucket rate limiter that applies to the probes of every stage.

- `probe_per_sec`: Sets the sustained probe rate, meaning how many tokens are added to the bucket each second.
- `probe_burst`: Sets the bucket capacity, meaning how many probes can run back to back before the limiter holds back further requests.

The combination of these two options controls the outgoing request rate independently of the worker count. This is useful for reducing the chance of hitting a rate limit or being detected by upstream network DPI.

Both values are validated:

- `probe_per_sec`: from 1 to 1,000,000
- `probe_burst`: from 1 to 10,000

A value outside the range is rejected, the default value replaces it, and the correction is recorded in the log. The effective default for each option depends on the platform. See the quick reference table for more information.

## Minimum probe duration

```toml
min_probe_duration = 50
```

Sets the minimum time each probe runs in milliseconds. If a probe finishes sooner than this time, the worker waits until the set minimum and then moves to the next target.

This option is useful for making probe times more even and for slowing scans that would otherwise finish very fast, including when you want to check timing based side channels.

The allowed range is from 10 milliseconds to 5 seconds. The effective default depends on the platform. See the quick reference table for the details.

## Related files

- [`icmp_settings.toml`](./icmp.md): ICMP scan settings
- [`tcp_settings.toml`](./tcp.md): TCP port scan settings
- [`http_settings.toml`](./http.md): HTTP, HTTPS, and HTTP/3 probe settings
- [`dns_settings.toml`](./dns.md): DNS scan settings
- [`xray_settings.toml`](./xray.md): Xray scan settings and outbound validation
- [`writer_settings.toml`](./writer.md): result saving settings
