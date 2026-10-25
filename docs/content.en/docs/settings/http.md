---
title: "HTTP Settings"
weight: 7
---

# HTTP settings

> 💡 **Note:** You do not need to edit the TOML file by hand. You can change these settings directly inside BGScan.
>
> Go to **Settings → HTTP Settings** from the main menu and set the options with the Inspector inside the TUI.

Settings file: `settings/http_settings.toml`

The HTTP probe sends one request to each target and records the status code, the agreed HTTP version, and whether TLS was used. HTTP/1.1 and HTTP/2 use TCP, but HTTP/3 runs over QUIC.

## Quick reference

| Setting | Default | Description |
|---|---|---|
| `workers` | Depends on the platform | The number of concurrent requests (1 to 1000) |
| `host` | `"example.com"` | The request host, with the option of setting a path |
| `server_name` | `""` | An alternate SNI value. When empty, `host` is used |
| `port` | `443` | The target port (1 to 65535) |
| `protocol` | `"https"` | The connection protocol: `http` or `https` |
| `version` | `"h1,h2"` | The HTTP version to use |
| `fingerprint` | `""` | The ClientHello fingerprint in uTLS. An empty value means using the Go standard library |
| `tls_validation` | `true` | Checks the validity of the server TLS certificate |
| `min_tls_version` | `"tls1.1"` | The lowest allowed TLS version |
| `max_tls_version` | `"tls1.3"` | The highest allowed TLS version |
| `timeout` | Depends on the platform | The request timeout in milliseconds (100 to 60000) |
| `accepted_status_codes` | `[]` | The acceptable status codes. An empty list means every code is accepted |
| `output_prefix` | `"http_"` | The result file name prefix |

## Workers

```toml
workers = 50
```

The number of requests that run at the same time. Each worker in the HTTP probe makes a connection and reads a response, so it usually uses more resources than a TCP scan worker.

Suggested values to start with:

- `10-50` for systems with limited resources
- `50-100` for typical use
- `100-200` for scanning with more capacity

## Host

```toml
host = "example.com"
```

The host the HTTP request is sent to. You can also set a path, for example `example.com/path`. This path is added to the request URL, but the `Host` header contains only the domain name.

## Port

```toml
port = 443
```

Sets the destination port of the request. The usual ports are:

- `80` for HTTP
- `443` for HTTPS

You can of course set any port the service listens on.

## Protocol

```toml
protocol = "https"
```

Sets the connection protocol. The value of this option must be `http` or `https`. Choosing `https` makes the connection use TLS.

## TLS validation

```toml
tls_validation = true
```

When this option is `true`, the server TLS certificate must be valid and trusted by the system.

With `false`, the certificate check is disabled and expired or self signed certificates are also accepted.

If you scan IPs directly, the server certificate may not match the destination IP. In this situation you can disable validation if needed.

## HTTP version

```toml
version = "h1,h2"
```

This option defines which HTTP version the probe uses.

| Value | Meaning |
|---|---|
| `h1` | HTTP/1.1 only |
| `h2` | HTTP/2 only |
| `h1,h2` | The version is chosen through ALPN |
| `h3` | HTTP/3 over QUIC |

These alternate names are also accepted:

- `http1`
- `http1.1`
- `http2`
- `http3`
- `http1,http2`
- `http2,http1`

HTTP/3 uses a separate probe over QUIC and always runs with TLS 1.3. The `min_tls_version` and `max_tls_version` settings therefore have no effect on it.

## Timeout

```toml
timeout = 4000
```

The maximum time assigned to each request, in milliseconds.

To start, you can consider these ranges:

- `2000-5000` for stable networks
- `5000-10000` for distant or unstable networks

If the timeout is too low, slow connections may be recorded as failed.

## TLS version range

```toml
min_tls_version = "tls1.1"
max_tls_version = "tls1.3"
```

These set the minimum and maximum TLS version that can be used in the handshake.

The allowed values for both options are:

- `tls1.0`
- `tls1.1`
- `tls1.2`
- `tls1.3`

The minimum version must not be higher than the maximum version. These settings are ignored for `version = "h3"`, because HTTP/3 always uses TLS 1.3.

## TLS fingerprint

```toml
fingerprint = ""
```

This option sets the ClientHello fingerprint using uTLS. When you give it a value, the probe uses the chosen profile instead of the default ClientHello of the Go standard library, so the handshake details are closer to the browser you have in mind.

This feature can be useful for checking how networks behave when they identify or filter traffic based on the TLS fingerprint.

Fingerprint names are not case sensitive.

| Category | Usable values |
|---|---|
| Chrome | `Chrome`, `Chrome_58`, `Chrome_62`, `Chrome_70`, `Chrome_72`, `Chrome_83`, `Chrome_87`, `Chrome_96`, `Chrome_100`, `Chrome_102`, `Chrome_120` |
| Firefox | `Firefox`, `Firefox_55`, `Firefox_56`, `Firefox_63`, `Firefox_65`, `Firefox_99`, `Firefox_102`, `Firefox_105`, `Firefox_120` |
| iOS | `iOS`, `iOS_11_1`, `iOS_12_1`, `iOS_13`, `iOS_14` |
| Other | `random` |

An empty value, meaning the default, uses the standard `crypto/tls` ClientHello in Go.

The HTTP/3 probe over QUIC does not use uTLS, so when `version = "h3"`, the chosen fingerprint is not applied.

## Server Name Indication (SNI)

```toml
server_name = ""
```

This option sets the SNI value sent during the TLS handshake.

If its value is empty, the domain name comes from `host`. If you want to test a specific IP with a particular domain name, or check Domain Fronting behavior, you can set the SNI value separately.

## Acceptable status codes

```toml
accepted_status_codes = []
```

Sets the list of HTTP status codes that count as a successful result.

If the list is empty or covers every status code, the filter is disabled and any status code the server returns is accepted.

For example, if you want only specific responses to be accepted, you can set it like this:

```toml
accepted_status_codes = [200, 204, 301, 302, 307, 308]
```

Responses whose status code is not in the list are not considered successful and do not reach the next stage of a multi stage scan.

## Output prefix

```toml
output_prefix = "http_"
```

Sets the file name prefix for the result files this probe produces. Files are saved under `result/http/`.

## Related files

- [`general_settings.toml`](./general.md): general scan settings and the scan execution mode
- [`icmp_settings.toml`](./icmp.md): ICMP scan settings
- [`tcp_settings.toml`](./tcp.md): TCP port scan settings
- [`dns_settings.toml`](./dns.md): DNS scan settings
- [`xray_settings.toml`](./xray.md): Xray scan settings and outbound validation
- [`writer_settings.toml`](./writer.md): result saving settings
