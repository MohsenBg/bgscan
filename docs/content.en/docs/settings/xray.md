---
title: "Xray Settings"
weight: 9
---

# Xray settings

> 💡 **Note:** You do not need to edit the TOML file by hand. You can change these settings directly inside BGScan.
>
> Go to **Settings → Xray Settings** from the main menu and set the options with the Inspector inside the TUI.

Settings file: `settings/xray_settings.toml`

This file defines the connection and speed test settings for Xray outbounds. Using this module, BGScan checks access to the destination servers and, when needed, the download and upload speed.

## Quick reference

| Setting | Default | Description |
|---|---|---|
| `timeout` | Depends on the platform | The maximum wait for the Xray connection test to finish, in milliseconds |
| `speed_test_timeout` | Depends on the platform | The maximum time for each download or upload speed test, in milliseconds |
| `workers` | Depends on the platform | The number of concurrent workers for the Xray connection test |
| `connectivity_test_type` | `3` | The test type: `0` connection only, `1` download only, `2` upload only, `3` download and upload |
| `download_speed` | Depends on the platform | The minimum acceptable download speed in Kbps |
| `upload_speed` | Depends on the platform | The minimum acceptable upload speed in Kbps |
| `output_prefix` | `"xray_"` | The output file name prefix for this module |
| `pre_scan_type` | `"tcp"` | The initial check before Xray: `none`, `icmp`, `tcp`, or `http` |
| `max_concurrent_dials` | Depends on the platform | The maximum number of concurrent outgoing dials in the whole process. A value of `0` means no limit |
| `dial_max_attempts` | Depends on the platform | The maximum number of dial attempts for each outbound dispatch |

## Connection test timeout

```toml
timeout = 6000
```

The maximum time the Xray connection test has to finish, in milliseconds. If the test does not finish in this time, it is considered failed.

**Suggested values:**

- `3000-6000` for stable networks
- `6000-12000` for distant or unstable networks

## Speed test timeout

```toml
speed_test_timeout = 6000
```

Sets the maximum time for each download or upload test in milliseconds.

The amount of data moved in the test is based on this time and the minimum speed. For download, the calculation is:

`(speed_test_timeout / 1000) × download_speed`

Upload uses `upload_speed` the same way. The result is in kilobits and is converted to bytes to calculate the data size. This way the test moves enough data to compare the connection speed with the set minimum.

**Suggested values:**

- `3000-8000` for a normal speed test
- `8000-15000` for high bandwidth targets or unstable networks

## Workers

```toml
workers = 32
```

Sets the number of workers that run the Xray connection test at the same time. Raising this value can increase test capacity, but CPU and bandwidth use go up too.

**Suggested values:**

- `8-16` for systems with limited resources
- `16-32` for typical use
- `32-64` for testing with more capacity

## Connection test type

```toml
connectivity_test_type = 0
```

Defines what kind of test BGScan runs on the outbound.

The supported values:

| Value | Test type | Description |
|---|---|---|
| `0` | Connection only | Checks whether a connection is made |
| `1` | Download only | Checks only the download speed |
| `2` | Upload only | Checks only the upload speed |
| `3` | Download and upload | Runs both speed tests |

## Minimum download speed

```toml
download_speed = 100
```

Sets the minimum acceptable download speed in Kbps. Based on `speed_test_timeout` and this value, the probe sets the amount of data needed for the test. If the measured speed is lower than the set minimum, the target is rejected.

This option is used only when `connectivity_test_type` is `1` or `3`.

**Suggested values:**

- `50-200` for testing moderate bandwidth
- `200-1000` for testing high bandwidth

## Minimum upload speed

```toml
upload_speed = 50
```

Sets the minimum acceptable upload speed in Kbps. This option works like `download_speed` and applies only when `connectivity_test_type` is `2` or `3`.

**Suggested values:**

- `25-100` for testing moderate bandwidth
- `100-500` for testing high bandwidth

## Output prefix

```toml
output_prefix = "xray_"
```

Sets the file name prefix for all output files from this module. By changing this value you can keep scan results that run with different targets or settings apart.

## Pre-scan type

```toml
pre_scan_type = "none"
```

Defines whether an initial reachability check runs before the Xray test.

If a target is rejected at this stage, it does not reach the Xray stage.

The usable values:

- `none`: No pre-scan. The Xray test runs directly.
- `icmp`: Reachability check with ICMP, using the ICMP settings.
- `tcp`: TCP connection check to the target port, using the TCP settings.
- `http`: Reachability check through HTTP or HTTPS, using the HTTP settings.

## Maximum concurrent dials

```toml
max_concurrent_dials = 128
```

Sets the maximum number of outgoing dials that can run at the same time in the whole process. This limit is shared across all workers.

A value of `0` means no limit is set for the number of concurrent dials.

If you run out of sockets or file descriptors during large scans, you can lower this value to limit the number of simultaneous connections.

The allowed range is `0` to `10000`.

## Maximum dial attempts

```toml
dial_max_attempts = 3
```

Sets how many times each outbound dispatch tries to make a connection before giving up.

The allowed range is `1` to `10`.

## Related files

- [`general_settings.toml`](./general.md): general scan settings and the scan execution mode
- [`icmp_settings.toml`](./icmp.md): ICMP scan settings
- [`tcp_settings.toml`](./tcp.md): TCP port scan settings
- [`http_settings.toml`](./http.md): HTTP, HTTPS, and HTTP/3 probe settings
- [`dns_settings.toml`](./dns.md): DNS scan settings
- [`writer_settings.toml`](./writer.md): result saving settings
