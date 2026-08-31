---
title: "IP Lists"
weight: 2
---

# IP Lists

Scanners read target addresses from IP list files. Before a list can be used for scanning, it must be prepared and imported into bgscan.

---

## Preparing a File to Import

The file you import must be a plain `.txt` file with **one entry per line**. Both **IPv4 and IPv6** are supported — single addresses and CIDR ranges of either family are accepted.

| Format | Example |
|---|---|
| Single IPv4 address | `192.168.1.1` |
| IPv4 CIDR range | `10.0.0.0/24` |
| Single IPv6 address | `2606:4700::6810:85e5` |
| IPv6 CIDR range | `2606:4700::/32` |

**Rules:**

- One entry per line — do not put multiple addresses on the same line.
- Empty lines and lines that cannot be parsed as a valid IPv4 or IPv6 address or CIDR are silently skipped.
- There is no comment syntax; lines starting with `#` will be skipped as unparseable.
- Leading/trailing whitespace on each line is trimmed automatically.

> **Note on large IPv6 ranges:** a `/32` or wider IPv6 prefix contains 2^96+ addresses. bgscan expands CIDRs on the fly during scanning, but scanning such a range to completion is impractical — use `max_ips_to_test` in [General Settings](../settings/general.md) to cap the count.

### Example import file

```
192.168.1.1
10.0.0.0/24
172.16.50.100
203.0.113.0/28
2606:4700::6810:85e5
2606:4700::/32
```

---

## Built-in IP Lists

bgscan ships with a set of pre-built IP lists in the `ips/` directory. These are automatically loaded and available for scanning without additional configuration.

| Name | Description |
|---|---|
| `akamai_IPv4` | Akamai CDN IPv4 ranges |
| `aws_IPv4` | Amazon Web Services IPv4 ranges |
| `azure_IPv4` | Microsoft Azure IPv4 ranges |
| `bunny_IPv4` | Bunny CDN IPv4 ranges |
| `cloudflare_IPv4` | Cloudflare IPv4 ranges |
| `cloudflare_IPv6` | Cloudflare IPv6 ranges |
| `cloudflare_ipv6_known` | Cloudflare IPv6 known-good ranges |
| `cloudflare_common_IPv6` | Cloudflare common IPv6 ranges |
| `cloudflare_full_IPv6` | Cloudflare full IPv6 ranges |
| `fastly_IPv4` | Fastly CDN IPv4 ranges |
| `gcore_IPv4` | G-Core Labs IPv4 ranges |
| `google_IPv4` | Google IPv4 ranges |
| `iran_IPv4` | Iranian IPv4 addresses |
| `pub_dns_IPv4` | Public DNS server IPv4 addresses |

These files follow the same internal CSV format as any imported list and can be renamed or deleted like any other file.

---

## Related Topics

- [Scan Source](./scan-source.md)
- [Scan Types](./scan-types.md)
- [Result Files](./result-files.md)
