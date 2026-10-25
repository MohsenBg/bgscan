---
title: "IP Lists"
weight: 2
---

# IP lists

The scanner reads target addresses from IP list files. To use a list, you must prepare it and import it into bgscan.

---

## Preparing a file to import

The input file must be a plain `.txt` file with **one entry per line**. Both IPv4 and IPv6 addresses are supported, along with CIDR ranges.

| Format | Example |
|---|---|
| IPv4 address | `192.168.1.1` |
| IPv4 CIDR range | `10.0.0.0/24` |
| IPv6 address | `2606:4700::6810:85e5` |
| IPv6 CIDR range | `2606:4700::/32` |

**Input file rules:**

- Each line must have only one entry. Do not write several addresses on one line.
- Empty lines and entries that are not a valid IPv4, IPv6, or CIDR address are skipped without showing an error.
- The file does not support comments. Even lines that start with `#` are skipped.
- Spaces at the start and end of each line are removed automatically.

> **About large IPv6 ranges:** A prefix such as ` /32` has more than 2^96 addresses. bgscan expands CIDR ranges during the scan, but scanning such a large range to completion is not practical. To limit the number of addresses that are checked, set the `max_ips_to_test` value in [General settings](../settings/general.md).

### Example input file

```
192.168.1.1
10.0.0.0/24
172.16.50.100
203.0.113.0/28
2606:4700::6810:85e5
2606:4700::/32
```

---

## Built-in lists

bgscan also has several ready IP lists in the `ips/` directory. These lists load automatically and need no extra settings.

| Name | Description |
|---|---|
| `akamai_IPv4` | Akamai IPv4 ranges |
| `aws_IPv4` | Amazon Web Services IPv4 ranges |
| `azure_IPv4` | Microsoft Azure IPv4 ranges |
| `bunny_IPv4` | Bunny CDN IPv4 ranges |
| `cloudflare_IPv4` | Cloudflare IPv4 ranges |
| `cloudflare_IPv6` | Cloudflare IPv6 ranges |
| `cloudflare_ipv6_known` | Healthy known Cloudflare IPv6 ranges |
| `cloudflare_common_IPv6` | Common Cloudflare IPv6 ranges |
| `cloudflare_full_IPv6` | All Cloudflare IPv6 ranges |
| `fastly_IPv4` | Fastly IPv4 ranges |
| `gcore_IPv4` | G-Core Labs IPv4 ranges |
| `google_IPv4` | Google IPv4 ranges |
| `iran_IPv4` | Iranian IPv4 addresses |
| `pub_dns_IPv4` | IPv4 addresses of public DNS servers |

The built-in lists use the same CSV format and you can rename or delete them like any other list.

---

## Related topics

- [Scan source](./scan-source.md)
- [Scan types](./scan-types.md)
- [Result files](./result-files.md)
