<div align="right">

[**English**](./http_settings.md)  |  [**فارسی**](./http_settings.fa.md)

</div>

# HTTP Scanner Configuration

Configuration file: `settings/http_settings.toml`

This file controls all HTTP and HTTPS-based scanning modules, allowing you to fine-tune transport security parameters, request timing, and concurrency levels.

---

## Table of Contents

- [HTTP — Connection Scanner](#http--connection-scanner)
  - [Core](#core)
  - [Timing & Concurrency](#timing--concurrency)
  - [TLS Settings](#tls-settings)
  - [Output](#output)

---

## `[http]` — Connection Scanner

Performs direct HTTP/HTTPS requests to target hosts to analyze network accessibility, web server headers, and cryptographic handshakes.

---

## Core

### `host`

```toml
host = "example.com"
```

Target host or URL path used for the HTTP request.

This value can be either a plain host or include a path.

| Value              | Description    |
| ------------------ | -------------- |
| `example.com`      | Plain host     |
| `example.com/path` | Host with path |

> Important: If a path is provided, it will be used in the request URL, but the `Host` header always uses only the domain part.

---

### `port`

```toml
port = 443
```

Target port used for the HTTP/HTTPS connection.

| Value | Description               |
| ----- | ------------------------- |
| `80`  | Standard HTTP connection  |
| `443` | Standard HTTPS connection |

---

### `protocol`

```toml
protocol = "https"
```

Application protocol used for the scan.

| Value     | Description              |
| --------- | ------------------------ |
| `"http"`  | Plaintext HTTP requests  |
| `"https"` | Encrypted HTTPS requests |

---

## Timing & Concurrency

### `timeout`

```toml
timeout = 4000
```

Timeout for each HTTP request in milliseconds. If the server does not respond within this time, the request is considered failed.

---

### `workers`

```toml
workers = 50
```

Number of concurrent worker goroutines performing HTTP scans.

> Recommended range: 10–150
> Higher values increase scanning speed but also increase CPU and network usage.

---

## TLS Settings

### `tls_validation`

```toml
tls_validation = true
```

Enable or disable TLS certificate validation when using HTTPS.

| Value   | Description                                                |
| ------- | ---------------------------------------------------------- |
| `true`  | Only valid and trusted certificates are accepted           |
| `false` | Self-signed, invalid, or expired certificates are accepted |

---

### `min_tls_version`

```toml
min_tls_version = "tls1.1"
```

Minimum TLS version allowed for HTTPS connections.

Common values:

- `tls1.0`
- `tls1.1`
- `tls1.2`
- `tls1.3`

---

### `max_tls_version`

```toml
max_tls_version = "tls1.3"
```

Maximum TLS version allowed for HTTPS connections.

---

### `server_name`

```toml
server_name = ""
```

Optional TLS Server Name Indication (SNI).

- Empty → automatically uses host domain
- Set value → overrides TLS server name

---

## Output

### `prefix_output`

```toml
prefix_output = "http_"
```

Prefix added to generated output file names (e.g., `http_results.txt`).

Useful when running multiple scanner types to easily distinguish outputs.

---

## Full Example

```toml
[http]
host            = "example.com"
port            = 443
protocol        = "https"
timeout         = 4000
workers         = 50
tls_validation  = true
min_tls_version = "tls1.1"
max_tls_version = "tls1.3"
server_name     = ""
prefix_output   = "http_"
```
