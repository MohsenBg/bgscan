---
title: "ICMP Settings"
weight: 5
---

# ICMP settings

> 💡 **Note:** You do not need to edit the TOML file by hand. You can change these settings directly inside BGScan.
>
> Go to **Settings → ICMP Settings** from the main menu and set the options with the Inspector inside the TUI.

Settings file: `settings/icmp_settings.toml`

The ICMP probe sends an Echo request to each target and records the round trip time of the reply. To receive replies it uses one shared IPv4 socket and, when possible, one shared IPv6 socket. Incoming replies are shared among the workers waiting for them.

## Quick reference

| Setting | Default | Range | Description |
|---|---|---|---|
| `timeout` | Depends on the platform | 100 to 30000 | The maximum wait for an Echo reply, in milliseconds |
| `tries` | `1` | 1 to 10 | The number of Echo requests for each target |
| `workers` | Depends on the platform | 1 to 5000 | The number of concurrent probes |
| `output_prefix` | `"icmp_"` | — | The result file name prefix |

## Timeout

```toml id="6c7k2a"
timeout = 2000
```

The maximum time each Echo request waits for a reply, in milliseconds.

This time is counted for each attempt separately, so in the worst case checking one target may take up to `timeout * tries`.

Suggested values to start with:

- `1500-2000` milliseconds for servers and desktops
- `2500-3000` milliseconds for Android devices

## Number of attempts

```toml id="k6j4ba"
tries = 1
```

The number of Echo requests sent for each target. If the target does not answer within the set attempts, it is considered unreachable.

The number of the attempt that received the reply is also saved in the result.

## Workers

```toml id="6fj7uk"
workers = 200
```

The number of targets checked at the same time.

The workers use shared sockets, so raising this value affects the packet rate more than the number of file descriptors used. Even so, a very high value may make intermediate routers limit requests or apply rate limits.

Suggested values to start with:

- `15-100` for Android devices
- `30-400` for desktops
- `100-1000` for servers

## Output prefix

```toml id="w4kx3d"
output_prefix = "icmp_"
```

Sets the file name prefix for the result files this probe produces. Files are saved under `result/icmp/`.

## Socket type

There is no separate setting for choosing the socket type, but the socket type used is recorded in each target result.

The probe first tries to open a Raw ICMP socket. If the program does not have the needed access, it falls back to a Datagram ICMP socket over UDP, which does not need special access.

The socket type used is recorded in the result with one of these values:

- `raw`
- `udp`

IPv6 support also depends on the system features and permissions. If the IPv6 socket cannot be opened, the IPv4 scan still runs, but IPv6 targets fail with an error.

## Related files

- [`general_settings.toml`](./general.md): general scan settings and the scan execution mode
- [`tcp_settings.toml`](./tcp.md): TCP port scan settings
- [`http_settings.toml`](./http.md): HTTP, HTTPS, and HTTP/3 probe settings
- [`dns_settings.toml`](./dns.md): DNS scan settings
- [`xray_settings.toml`](./xray.md): Xray scan settings and outbound validation
- [`writer_settings.toml`](./writer.md): result saving settings
