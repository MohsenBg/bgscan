---
title: "Settings Overview"
weight: 1
---

# Settings overview

BGScan settings are kept in TOML files inside the `settings/` directory. You can change most scan and interface settings from inside the program too. Changes made through the Inspector are saved to disk automatically.

## Settings files

| File | Purpose |
|---|---|
| `settings/general_settings.toml` | General scan settings and the scan execution mode |
| `settings/writer_settings.toml` | Buffer settings and saving results to disk |
| `settings/icmp_settings.toml` | ICMP probe settings |
| `settings/tcp_settings.toml` | TCP connection settings |
| `settings/http_settings.toml` | HTTP, HTTPS, and HTTP/3 probe settings |
| `settings/xray_settings.toml` | Xray scan settings |
| `settings/dns_settings.toml` | DNS scan and DNS tunnel settings |
| `settings/ui_settings.toml` | Theme, background, animation, maximum table rows, and log level settings |

## Two ways to change settings

- **Editing the TOML files:** Open the file you want from the `settings/` directory with a text editor, change the values, and save. Run BGScan again for the changes to take effect. This method also works well for managing settings with version control.
- **Using the Inspector inside the program:** Run BGScan, go to **Settings** from the main menu, and choose the section you want. Pressing `Enter` on any option lets you change its value. Changes are saved to disk right away. Interface settings such as the theme and log level also take effect without a restart.

The Inspector shows only the options related to the current settings. For example, the HTTP TLS options appear only when HTTPS is selected.

## Default values and validation

The default values are defined inside BGScan itself. If a settings file does not exist, the program creates it with the default values on the first run. Fields that are missing from a file also get their own default value.

However, if a file exists and its content cannot be parsed, the program reports an error and does not silently go back to the default values.

When the program starts, the loaded settings are validated. If a value is outside the allowed range, it is replaced with the default value and this is reported in the initial program checks. An invalid value therefore usually does not stop the program from running.

Corrected values change only in memory at first. To save them to the file, you must save that section through the Inspector.

## DNS tunnel settings

The settings for the DNS tunnel protocols, including DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed, are stored in separate TOML files inside the `assets/dns-tunneling/` directory, not in `dns_settings.toml`.

To create and manage these settings, go to this path inside the program:

**Main Menu → DNS Tunneling**
