---
title: "TheFeed"
weight: 6
---

# TheFeed settings

TheFeed runs inside BGScan itself through an internal client library. Queries and responses are encrypted with AES-256 keys built from a shared passphrase.

This protocol has no proxy settings. The probe timeout also comes from the general DNS tunnel settings, meaning `dns_settings.toml` and the `dns_tunneling.timeout` option, so you do not need to set it in the TheFeed config file.

## Connection settings

| Field | Default | Description |
| --- | --- | --- |
| `domain` | Required | The subdomain that answers Feed queries, for example `t.example.com` |
| `passphrase` | Required | The same server passphrase used to build the AES-256 keys |
| `resolver_port` | `53` | The resolver port. TheFeed servers usually use port `5300` |
| `resolver_type` | `'udp'` | The transport for talking to the resolver: `udp`, `tcp`, or `dot` |
| `query_mode` | `'single'` | The query encoding method: `single` for one Base32 label or `double` for several Hex labels |

## Example config

```toml id="q6v3nw"
domain = 't.example.com'
passphrase = 'your-passphrase'
resolver_port = 53
resolver_type = 'udp'
query_mode = 'single'
```

## Related topics

- [DNS tunnel](../): an introduction to the protocols and the scan flow
- [DNS settings](../settings/dns.md): the tunnel timeout and resolver pre-scan settings
