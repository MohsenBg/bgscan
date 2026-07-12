---
title: "IP Files"
weight: 4
---

# IP Files

Scanners read target addresses from IP list files. Before a list can be used for scanning it must be imported through the TUI, which validates and converts it into bgscan's internal format.

---

## Preparing a File to Import

The file you import must be a plain `.txt` file with **one entry per line**. Only **IPv4** is supported — IPv6 addresses and prefixes are not accepted and will be skipped on import.

| Format | Example |
|---|---|
| Single IPv4 address | `192.168.1.1` |
| IPv4 CIDR range | `10.0.0.0/24` |

**Rules:**

- One entry per line — do not put multiple addresses on the same line.
- Empty lines and lines that cannot be parsed as a valid IPv4 address or CIDR are silently skipped.
- There is no comment syntax; lines starting with `#` will be skipped as unparseable.
- Leading/trailing whitespace on each line is trimmed automatically.

#### Example import file

```
192.168.1.1
192.168.1.2
10.0.0.0/24
172.16.50.100
203.0.113.0/28
```

---

## Default IP Lists

bgscan ships with a set of pre-built IP lists in the `ips/` directory. These are ready to use without any import step:

| Name | Description |
|---|---|
| `akamai` | Akamai CDN IPv4 ranges |
| `aws` | Amazon Web Services IPv4 ranges |
| `azure` | Microsoft Azure IPv4 ranges |
| `bunny` | Bunny CDN IPv4 ranges |
| `cloudflare` | Cloudflare IPv4 ranges |
| `fastly` | Fastly CDN IPv4 ranges |
| `gcore` | G-Core Labs IPv4 ranges |
| `google` | Google IPv4 ranges |
| `iran` | Iranian IPv4 addresses |

These files follow the same internal CSV format as any imported list and can be renamed or deleted like any other file.

---

## Managing IP Lists

Open the IP list manager from **Main Menu → IP List**.

The table lists every imported file sorted newest-first, showing its name, import date, and size. Three keyboard actions are available:

| Key | Action |
|---|---|
| `a` | Add — import a new `.txt` file |
| `r` | Rename — rename the selected file |
| `x` | Delete — permanently remove the selected file |

{{< img "/bgscan-iplist.webp" "bgscan iplist" >}}

#### Adding a file (`a`)

1. Press `a`. A file picker opens.
2. Navigate to and select your `.txt` file.
3. A name prompt appears. Type a name for this list (e.g. `internal-servers`).
   - The name must be a valid filename.
   - **The name must be unique** — you cannot use the same name as an existing list.
4. Press Enter to confirm. bgscan copies and converts the file into the `ips/` directory.

During import every line is parsed and normalised. Invalid lines are dropped silently; they do not cause the import to fail.

{{< img "/bgscan-select-iplist.webp" "bgscan select iplist" >}}

#### Renaming a file (`r`)

Select a file in the table and press `r`. Enter the new name at the prompt and press Enter. The file on disk is renamed immediately.

> The new name must also be unique — you cannot rename a file to a name that is already in use.

#### Deleting a file (`x`)

Select a file and press `x`. The file is removed from the `ips/` directory permanently. There is no undo.

---

## Internal Storage Format

After import, bgscan stores IP lists as CSV files in the `ips/` directory next to the binary:

```
<bgscan-root>/
└── ips/
    ├── internal-servers.csv
    ├── cloudflare.csv
    └── ...
```

Each `.csv` file uses a two-column format:

```
<ip_or_cidr>,<enable>
```

- `enable` is `1` (active) or `0` (disabled). Entries you imported from a `.txt` file are always written as `1`.
- When the enable column is absent, bgscan treats the entry as enabled.
- You do not need to hand-edit these files. The TUI manages them for you.

The file's display name in the TUI is the filename without the `.csv` extension.

---

## Related Topics

- [Scanner Overview](../scanner.md)
- [Scan Types](./scan-types.md)
- [Result Files (Output)](./result-files.md)
- [Scan Pipeline](./scan-pipeline.md)
