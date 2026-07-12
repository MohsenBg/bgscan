---
title: "Settings Overview"
weight: 1
---

# Settings Overview

bgscan keeps all configuration in plain TOML files inside `settings/`. The app also exposes these same options through the in-app Settings inspector, which writes changes back to disk automatically.

## Setting files

| File | Purpose |
|------|---------|
| `settings/general_settings.toml` | Global scan control and pipeline mode |
| `settings/writer_settings.toml` | Result buffering and disk writes |
| `settings/icmp_settings.toml` | ICMP probe tuning |
| `settings/tcp_settings.toml` | TCP connect tuning |
| `settings/http_settings.toml` | HTTP/HTTPS/HTTP3 probe tuning |
| `settings/xray_settings.toml` | Xray test tuning |
| `settings/dns_settings.toml` | DNS resolver, DNSTT, and SlipStream tuning |

## Two ways to edit settings

- **TOML files** — open any `settings/*.toml` file in any text editor, change values, save, and restart bgscan. This is the most explicit approach and works well for version control.
- **In-app inspector** — open bgscan, navigate to **Settings** in the main menu, pick a category, press `Enter` on any field to edit it. Changes are saved immediately to disk.

Some fields are dynamic. For example, TLS-related options only show when HTTPS is selected, and DNSTT/Slipstream fields only show when those probes are enabled.

## Default values

Defaults ship in bgscan.  They are copied into place by the installer or used as fallback values when a TOML file is missing or a field is absent.
