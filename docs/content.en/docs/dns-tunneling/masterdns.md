---
title: "MasterDNS"
weight: 4
---

# MasterDNS Configuration

MasterDNS runs in-process through a vendored client library. The tunnel negotiates MTU sizes automatically and protects traffic with a configurable encryption method. There are no proxy settings; the probe dials the resolver directly.

## Connection Settings

| Field | Default | Description |
| --- | --- | --- |
| `Domain` | (required) | Zone served by your MasterDNS server |
| `DataEncMethod` | `1` | Encryption: `0` = None, `1` = XOR (default), `2` = ChaCha20, `3` = AES-128-GCM, `4` = AES-192-GCM, `5` = AES-256-GCM |
| `EncryptionKey` | (required unless method is None) | 32 hex characters, must match the server |
| `ResolverPort` | `53` | Resolver port |

## MTU and Session Tuning

| Field | Default | Description |
| --- | --- | --- |
| `MTUTestTimeoutSec` | `3.0` | Seconds to wait per MTU probe packet |
| `MTUTestRetries` | `2` | Retries per MTU size before trying the next |
| `SessionInitRetryMaxSec` | `2.0` | Retry window for session init before restarting |
| `SessionInitRacingCount` | `1` | Parallel session init attempts |
| `MinUploadMTU` / `MaxUploadMTU` | `38` / `150` | Upload MTU range in bytes |
| `MinDownloadMTU` / `MaxDownloadMTU` | `100` / `500` | Download MTU range in bytes |
| `MTUParallelism` | `1` | Parallel MTU test workers |
| `RxTxWorkers` | `4` | Concurrent tunnel data streams |

## Example Config

```toml
domain = 'example.com'
data_enc_method = 1
encryption_key = 'cd6d78e954f48f62cb74cdcf8a2459d3'
resolver_port = 53
mtu_test_timeout_sec = 2.0
mtu_test_retries = 2
session_init_retry_max_sec = 60.0
session_init_racing_count = 3
min_upload_mtu = 38
max_upload_mtu = 150
min_download_mtu = 100
max_download_mtu = 500
mtu_parallelism = 3
rx_tx_workers = 2
```

## Related Topics

- [DNS Tunneling](../) — Protocol overview and scan flow
- [StormDNS](../stormdns/) — Same model with its own defaults and query type selection
