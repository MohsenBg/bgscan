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

- **TOML files** — open any `settings/*.toml` file in a text editor, change values, save, and restart bgscan. This works well under version control.
- **In-app inspector** — open bgscan, navigate to **Settings** in the main menu, pick a category, press `Enter` on any field to edit it. Changes are saved immediately to disk.

Some fields are dynamic. For example, TLS-related options only show when HTTPS is selected, and DNSTT/Slipstream fields only show when those probes are enabled.

## Defaults and validation

Defaults are compiled into bgscan. A missing settings file is created with its defaults on first run, and a field absent from an existing file falls back to its default. A file that exists but cannot be parsed is an error, not a fallback.

On startup, every loaded config is validated. Out-of-range values are replaced with the default and the correction is logged, so a bad edit degrades to the default instead of failing the run.

The `settings/*.toml.default` files in the repository are reference copies of the shipped defaults. bgscan does not read them at runtime.
