<div align="left">

[**English**](./README.md)  |  [**فارسی**](./README.fa.md)

</div>

# bgscan — User Guide

bgscan is a terminal-based tool for scanning IP lists and ranges using
multiple scan types (ICMP, TCP, DNS, Xray) and reviewing the results.

This guide walks through the main menu and how to use each section.

## Table of Contents

- [Main Menu](#main-menu)
- [Run Scan](#run-scan)
  - [Changing scan settings](#changing-scan-settings)
- [IPList](#iplist)
  - [Managing lists](#managing-lists)
  - [Adding your own list](#adding-your-own-list)
- [ResultList](#resultlist)
- [Xray Outbound](#xray-outbound)
- [Log](#log)

## Main Menu

The app is organized into five sections:

- **Run Scan** — start a new scan
- **IPList** — manage the IP lists / ranges you scan
- **ResultList** — view results from previous scans
- **Xray Outbound** — manage Xray outbound configs (used by the Xray scan type)
- **Log** — view UI, core, and debug logs

---

## Run Scan

To run a scan:

1. Select an **IP list** as your scan source. This can be:
   - An IP list you've already added (see [IPList](#iplist)), or
   - A result file from a previous scan (see [ResultList](#resultlist))
2. Select a **scan type**: `icmp`, `tcp`, `dns`, or `xray`.
   - If you choose `xray`, you'll also need to select an **outbound**
     (see [Xray Outbound](#xray-outbound)).
3. Start the scan.

### Changing scan settings

Each scan type has its own settings file under `settings/*.toml`. Edit the
relevant file to change how that scan type behaves.

See the settings docs for each scan type:

- [General Settings](general_settings.md)
- [ICMP Settings](icmp_settings.md)
- [TCP Settings](tcp_settings.md)
- [DNS Settings](dns_settings.md)
- [HTTP Settings](http_settings.md)
- [Xray Settings](xray_settings.md)
- [Writer Settings](writer_settings.md)

---

## IPList

This menu shows the IP lists available to scan. Built-in lists include:

```
akamai_range_IPv4.csv  azure_range_IPv4.csv  cloudflare_range_IPv4.csv
aws_range_IPv4.csv     bunny_range_IPv4.csv  fastly_range_IPv4.csv
gcore_range_IPv4.csv   google_range_IPv4.csv iran_range_IPv4.csv
```

> **Note:** Only IPv4 is currently supported. IPv6 support is planned.

### Managing lists

| Key | Action |
|-----|--------|
| `r` | Rename the selected list |
| `x` | Delete the selected list |
| `a` | Add a new list from a text file |


### Adding your own list

When adding a list, the file should contain one IP or IP range per line.
Single IPs and CIDR ranges can be mixed freely in the same file.

**Example 1 — single IPs:**
```
192.168.1.1
192.168.1.10
192.168.1.20
```

**Example 2 — ranges:**
```
192.168.1.1/24
192.168.2.1/24
```

**Example 3 — mixed:**
```
192.168.1.1
192.168.2.1/24
```

---

## ResultList

This menu stores the results of your scans. Select a result file to view
the IPs it found.

| Key     | Action |
|---------|--------|
| `Enter` | Copy the selected IP |
| `r`     | Rename the selected result file |
| `x`     | Delete the selected result file |

---

## Xray Outbound

This menu manages your Xray outbound configurations, used when running an
Xray scan. For details on outbound formats and how to add your own, see:

- [Xray Outbounds](xray_outbounds.md)

| Key | Action |
|-----|--------|
| `r` | Rename the selected outbound |
| `x` | Delete the selected outbound |
| `a` | Add a new outbound |



---

## Log

Three log types are available:

- **UI log** — events from the interface
- **Core log** — scanner log, where scan activity is recorded
- **Debug log** — intended for developers

If you want to see what's happening during a scan, watch the core log.
You can also view the scanner log live, during a scan, by pressing `l`.
