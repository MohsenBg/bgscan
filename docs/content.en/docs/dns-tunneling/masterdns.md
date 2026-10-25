---
title: "MasterDNS"
weight: 4
---

# MasterDNS settings

MasterDNS runs inside BGScan itself using an internal client library. The tunnel negotiates the MTU size automatically and protects traffic with the chosen encryption method.

This protocol has no proxy settings and the probe connects directly to the resolver.

## Connection settings

| Field | Default | Description |
| --- | --- | --- |
| `domain` | Required | The domain your MasterDNS server serves |
| `data_enc_method` | `1` | The encryption method: `0` = none, `1` = XOR (default), `2` = ChaCha20, `3` = AES-128-GCM, `4` = AES-192-GCM, `5` = AES-256-GCM |
| `encryption_key` | Required, except in the unencrypted mode | A 32 character hex key that must match the server key |
| `resolver_port` | `53` | The resolver port |

## MTU and session settings

| Field | Default | Description |
| --- | --- | --- |
| `mtu_test_timeout_sec` | `3.0` | How long to wait for each MTU test packet, in seconds |
| `mtu_test_retries` | `2` | The number of attempts for each MTU size before moving to the next size |
| `session_init_retry_max_sec` | `2.0` | How long to retry bringing up the session before starting over |
| `session_init_racing_count` | `1` | The number of parallel attempts to bring up the session |
| `min_upload_mtu` / `max_upload_mtu` | `38` / `150` | The minimum and maximum upload MTU size in bytes |
| `min_download_mtu` / `max_download_mtu` | `100` / `500` | The minimum and maximum download MTU size in bytes |
| `mtu_parallelism` | `1` | The number of concurrent workers for the MTU test |
| `rx_tx_workers` | `4` | The number of concurrent data streams in the tunnel |

## Example config

```toml id="j3p9xv"
domain = 'example.com'
data_enc_method = 1
encryption_key = 'cd6d78e954f48f62cb74cdcf8a2459d3'
resolver_port = 53
mtu_test_timeout_sec = 3.0
mtu_test_retries = 2
session_init_retry_max_sec = 2.0
session_init_racing_count = 1
min_upload_mtu = 38
max_upload_mtu = 150
min_download_mtu = 100
max_download_mtu = 500
mtu_parallelism = 1
rx_tx_workers = 4
```

## Related topics

- [DNS tunnel](../): an introduction to the protocols and the scan flow
- [StormDNS](./stormdns/): a protocol with a similar structure, its own default settings, and its own query type options
