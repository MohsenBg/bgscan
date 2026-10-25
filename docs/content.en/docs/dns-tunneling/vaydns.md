---
title: "VayDNS"
weight: 2
---

# VayDNS settings

VayDNS is the native version of the VayDNS protocol. It uses the same stack as DNSTT (DNS, then KCP, then Noise, then smux), but has its own framing and also gives you several options to set the QNAME structure, the MTU value, and the record type.

## Connection settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `domain` | Required | Valid domain | The zone delegated to your VayDNS server |
| `pub_key` | Required | 64 hex characters | The server public key |
| `resolver_type` | `"udp"` | `udp`, `tcp`, `dot` | The connection type to the resolver |
| `resolver_port` | `53` | Greater than `0` | The resolver port |
| `fingerprint` | `"Chrome"` | The DNSTT fingerprints | The ClientHello fingerprint in uTLS for DNS-over-TLS |
| `record_type` | `"TXT"` | The table below | The record type for tunnel queries |
| `rps` | `0` | `0` to `500` | The query rate limit per second. A value of `0` means no limit |

## Advanced settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `client_id_size` | `2` | `1` to `8` | The Client ID size in bytes |
| `max_qname_len` | `101` | `0` to `253` | The maximum QNAME length in DNS packets. A value of `0` means automatic |
| `max_num_labels` | `0` | `0` to `4` | The maximum number of labels in the QNAME. A value of `0` means automatic |
| `mtu` | `0` | `0` to `1452` | The maximum transmission unit. A value of `0` means automatic |

## Record types

The allowed values for `record_type` are: `A`, `AAAA`, `CNAME`, `NS`, `MX`, `TXT`, `SRV`, `NULL`, and `CAA`.

This option defines which DNS record type the tunnel queries use. The default is `TXT`, which has the widest support.

## Proxy and authentication

VayDNS uses the same proxy and authentication settings as DNSTT. See the [DNSTT proxy and authentication](./dnstt/#proxy-and-authentication) section for the details.

## Example config

```toml
domain = "example.com"
pub_key = "0000000000000000000000000000000000000000000000000000000000000000"
resolver_type = "udp"
resolver_port = 53
fingerprint = "Chrome"
record_type = "TXT"
client_id_size = 2
max_qname_len = 101
max_num_labels = 0
mtu = 0
rps = 0
proxy_type = "socks"
proxy_port = 1080
auth_method = "none"
```

## Related topics

- [DNS tunnel](../): an introduction to the protocols and the scan flow
- [DNSTT](./dnstt/): an introduction to DNSTT compatible framing and TLS fingerprints
