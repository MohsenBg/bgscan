---
title: "StormDNS"
weight: 5
---

# StormDNS Configuration

StormDNS runs in-process through a vendored client library. It works like MasterDNS with its own defaults, plus a selectable DNS query type used to carry the tunnel traffic. There are no proxy settings.

## Connection Settings

| Field | Default | Description |
| --- | --- | --- |
| `Domain` | (required) | Zone served by your StormDNS server |
| `DataEncMethod` | `1` | Encryption: `0` = None, `1` = XOR (default), `2` = ChaCha20, `3` = AES-128-GCM, `4` = AES-192-GCM, `5` = AES-256-GCM |
| `EncryptionKey` | (required unless method is None) | 32 hex characters, must match the server |
| `ResolverPort` | `53` | Resolver port |
| `DNSQueryType` | `'TXT'` | Query type carrying the tunnel: `TXT`, `NS`, `CNAME`, or `ROTATE` (empty means TXT) |

## MTU and Session Tuning

| Field | Default | Description |
| --- | --- | --- |
| `MTUTestTimeoutSec` | `2.0` | Seconds to wait per MTU probe packet |
| `MTUTestRetries` | `3` | Retries per MTU size before trying the next |
| `SessionInitRetryMaxSec` | `60.0` | Retry window for session init before restarting |
| `MinUploadMTU` / `MaxUploadMTU` | `100` / `200` | Upload MTU range in bytes |
| `MinDownloadMTU` / `MaxDownloadMTU` | `1000` / `4000` | Download MTU range in bytes |
| `MTUParallelism` | `16` | Parallel MTU test workers |
| `RxTxWorkers` | `4` | Concurrent tunnel data streams |

## Example Config

```toml
domain = 'example.com'
data_enc_method = 1
encryption_key = 'cd6d78e954f48f62cb74cdcf8a2459d3'
resolver_port = 53
dns_query_type = 'TXT'
mtu_test_timeout_sec = 2.0
mtu_test_retries = 3
session_init_retry_max_sec = 60.0
min_upload_mtu = 100
max_upload_mtu = 200
min_download_mtu = 1000
max_download_mtu = 4000
mtu_parallelism = 16
rx_tx_workers = 4
```

## Related Topics

- [DNS Tunneling](../) — Protocol overview and scan flow
- [MasterDNS](../masterdns/) — Same model with its own defaults
