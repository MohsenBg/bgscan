---
title: "DNSTT"
weight: 1
---

# DNSTT settings

DNSTT uses the VayDNS library with DNSTT compatible framing, meaning DNS-over-TXT with old style framing. The tunnel structure is: DNS packet channel, then KCP, then Noise encryption, then smux multiplexing.

## Connection settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `domain` | Required | Valid domain | The zone delegated to your DNSTT server |
| `pub_key` | Required | 64 hex characters | The server public key |
| `resolver_type` | `"udp"` | `udp`, `tcp`, `dot` | The transport for talking to the resolver |
| `resolver_port` | `53` | Greater than `0` | The resolver port |
| `fingerprint` | `"Chrome"` | According to the table below | The ClientHello fingerprint in uTLS for DNS-over-TLS |
| `rps` | `0` | 0 to 500 | The query rate limit per second. A value of `0` means no limit |

## Proxy and authentication

| Field | Default | Description |
| --- | --- | --- |
| `proxy_type` | `"socks"` | The proxy type: `socks` or `ssh` |
| `proxy_port` | `1080` | The proxy port |
| `auth_method` | `"none"` | The authentication method: `none`, `password`, or `key` |
| `username` | `""` | The user name. Needed for Password or Key authentication |
| `password` | `""` | The password. Needed for Password authentication |
| `private_key` | `""` | The SSH private key encoded in PEM. Needed for Key authentication |
| `known_hosts_file` | `""` | The path to the SSH `known_hosts` file. Needed for `proxy_type = "ssh"` and validated at run time |

**Proxy notes:**

- The SSH proxy needs Password or Key authentication and a `known_hosts_file` file. BGScan does not automatically look for `~/.ssh/known_hosts`, and this field is not shown in the TUI form. To use SSH, you must enter the file path directly in the TOML file.
- The SOCKS proxy does not support Key authentication.

## TLS fingerprints

The `fingerprint` field defines which ClientHello profile in uTLS is used during the TLS handshake. Names are compared without regard to upper or lower case.

| Category | Usable values |
| --- | --- |
| Chrome | `Chrome`, `Chrome_58`, `Chrome_62`, `Chrome_70`, `Chrome_72`, `Chrome_83`, `Chrome_87`, `Chrome_96`, `Chrome_100`, `Chrome_102`, `Chrome_120` |
| Firefox | `Firefox`, `Firefox_55`, `Firefox_56`, `Firefox_63`, `Firefox_65`, `Firefox_99`, `Firefox_102`, `Firefox_105`, `Firefox_120` |
| iOS | `iOS`, `iOS_11_1`, `iOS_12_1`, `iOS_13`, `iOS_14` |
| Other | `random` |

## Example config

```toml
domain = "example.com"
pub_key = "0000000000000000000000000000000000000000000000000000000000000000"
resolver_type = "udp"
resolver_port = 53
fingerprint = "Chrome"
rps = 0
proxy_type = "socks"
proxy_port = 1080
auth_method = "none"
```

## Related topics

- [DNS tunnel](../): an introduction to the protocols and the scan flow
- [VayDNS](./vaydns/): the native VayDNS protocol with similar proxy options
