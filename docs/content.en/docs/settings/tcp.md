---
title: "TCP Settings"
weight: 6
---

# TCP settings

> 💡 **Note:** You do not need to edit the TOML file by hand. You can change these settings directly inside BGScan.
>
> Go to **Settings → TCP Settings** from the main menu and set the options with the Inspector inside the TUI.

Settings file: `settings/tcp_settings.toml`

The TCP probe connects to a specific port on each target and records the handshake time. At this stage no application protocol connection is made.

## Quick reference

| Setting | Default | Range | Description |
|---|---|---|---|
| `port` | `443` | 1 to 65535 | The destination TCP port |
| `timeout` | `2000` | 100 to 30000 | The maximum wait for a connection, in milliseconds |
| `workers` | Depends on the platform | 1 to 5000 | The number of concurrent connection attempts |
| `tries` | `1` | 1 to 10 | The number of attempts for each target |
| `output_prefix` | `"tcp_"` | — | The result file name prefix |

## Port

```toml
port = 80
```

The TCP port checked on each target. Common ports are:

- `80` for HTTP
- `443` for HTTPS
- `22` for SSH

You can of course enter any port you want to check.

## Timeout

```toml
timeout = 3000
```

The maximum time the probe waits to make the connection and complete the handshake, in milliseconds. If the connection is not made within this time, the attempt is considered failed.

Suggested values to start with:

- `1000-3000` milliseconds for local networks
- `5000-10000` milliseconds for distant or unstable networks

## Workers

```toml
workers = 200
```

The number of connections checked at the same time. Each worker opens its own socket, so as you raise the worker count you should also consider the system file descriptor limit.

Suggested values to start with:

- `50-100` for systems with limited resources
- `200-500` for typical use
- `500-1000` for scanning with more capacity

## Number of attempts

```toml
tries = 1
```

The number of attempts made for each target. On a timeout the connection is tried again, but immediate errors such as `connection refused` make the target fail right away. This behavior helps scans of ranges whose ports are mostly closed finish faster.

The minimum value for this option is `1`, so the first attempt cannot be disabled.

## Output prefix

```toml
output_prefix = "tcp_"
```

Sets the file name prefix for the result files this probe produces. Files are saved under `result/tcp/`.

## Related files

- [`general_settings.toml`](./general.md): general scan settings and the scan execution mode
- [`icmp_settings.toml`](./icmp.md): ICMP scan settings
- [`http_settings.toml`](./http.md): HTTP, HTTPS, and HTTP/3 probe settings
- [`dns_settings.toml`](./dns.md): DNS scan settings
- [`xray_settings.toml`](./xray.md): Xray scan settings and outbound validation
- [`writer_settings.toml`](./writer.md): result saving settings
