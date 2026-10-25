---
title: "Xray Outbounds"
weight: 6
---

# Xray outbounds

Each outbound defines how BGScan passes traffic through the proxy during an Xray scan. Every outbound is saved as a JSON file in the `assets/xray/outbounds/` directory and defines the proxy protocol, transport, and security settings.

To run an Xray scan, you need at least one outbound.

To manage the outbounds, go to **Xray Outbound** from the main menu (`x`).

---

{{< img "/bgscan-outbound-menu.webp" "bgscan outbound menu" >}}

## Outbound list

The table shows the available outbounds from newest to oldest:

| Column | Description |
|---|---|
| Name | The outbound name without the `.json` extension |
| Protocol | The proxy protocol, such as `vless`, `vmess`, or `socks` |
| Network | The transport type, such as `ws`, `grpc`, `xhttp`, or `tcp` |
| Security | The `streamSettings.security` value, such as `tls`, `reality`, or an empty value |
| Created | The time the file was last changed |

Use these keys to manage the outbounds:

| Key | Action |
|---|---|
| `a` | Add a new outbound |
| `r` | Rename the selected outbound |
| `x` | Delete the selected outbound |

The sort and filter keys are also available:

- `s`: change the sort column
- `o`: reverse the sort order
- `/`: open the filter
- `b` or `Esc`: go back to the previous screen

---

{{< img "/bgscan-outbound-empty.webp" "bgscan outbound empty" >}}

## Adding an outbound (`a`)

Pressing `a` opens the **Addition Method** menu. There are two ways to add an outbound.

---

### First method: From Link

Choose **From Link**, enter the proxy link, and then set a name for the outbound. BGScan parses the link, extracts the outbound details from it, and saves them as a JSON file.

{{< img "/bgscan-outbound-select-link.webp" "bgscan outbound select link" >}}

These links are supported:

| Link scheme | Protocol |
|---|---|
| `vmess://` | VMess |
| `vless://` | VLESS |
| `trojan://` | Trojan |
| `ss://` | Shadowsocks |
| `socks://` | SOCKS |
| `hysteria2://` or `hy2://` | Hysteria2 |

WireGuard share links are not supported, but you can import the related outbound through a JSON file.

After entering the link, set the outbound name. This name must be unique and must not already be used for another outbound.

```text
a → From Link → Paste link → Enter name → Done
```

{{< img "/bgscan-outbound-link.webp" "bgscan outbound link" >}}
{{< img "/bgscan-outbound-name.webp" "bgscan outbound name" >}}

---

### Second method: From File

Choose **From File** and open a JSON or JSONC file from your system. BGScan extracts the proxy outbound from the file, checks that it is valid, and saves it.

{{< img "/bgscan-outbound-select-json.webp" "bgscan outbound select json" >}}

The input file can be one of these forms:

- **A single outbound:** For example, a JSON object with a `protocol` field.
- **An array of outbounds:** Containing several outbounds.
- **A full Xray config:** A file that has an `outbounds` array.

If the file has several outbounds, BGScan chooses the first outbound it recognizes as a proxy. If none is recognized that way, it chooses the first outbound with the tag `proxy`. The rest of the full config is not saved.

**Input file conditions:**

- The file must be valid JSON or JSONC. The `//` and `/* */` comments and a UTF-8 BOM are also accepted.
- The outbound must have a destination address that BGScan can replace with each target IP. This address can be in `settings.address`, one of the `settings.vnext` or `settings.servers` entries, or the endpoint of a WireGuard peer.
- The proxy protocol must be valid.
- The config must pass the internal Xray core validation. BGScan does not use an external `xray` binary for this.

You can use `$ADDRESS` as an address placeholder, but it is not required. Any address in one of the replaceable fields is swapped for the target IP during the scan, and the outbound port is kept. So if you already have a ready config, you do not need to remove the original server address before importing it.

A minimal example of a valid outbound:

```json
{
  "protocol": "vless",
  "settings": {
    "vnext": [
      {
        "address": "$ADDRESS",
        "port": 443,
        "users": [
          { "id": "your-uuid-here", "encryption": "none" }
        ]
      }
    ]
  },
  "streamSettings": {
    "network": "ws",
    "security": "tls",
    "tlsSettings": { "serverName": "example.com" },
    "wsSettings": { "path": "/ws", "headers": { "Host": "example.com" } }
  }
}
```

After choosing the file, enter a name for the outbound. This name must also be unique.

```text
a → From File → Select file → Enter name → Done
```

{{< img "/bgscan-outbound-name.webp" "bgscan outbound name" >}}
{{< img "/bgscan-outbound-item.webp" "bgscan outbound item" >}}

---

## Where outbounds are stored

Outbounds are saved as `.json` files in this directory:

```text
<bgscan-root>/
└── assets/
    └── xray/
        └── outbounds/
            ├── my-vless-ws.json
            ├── cloudflare-trojan.json
            └── ...
```

The name you see in the TUI is the file name without the `.json` extension.

BGScan also places several example outbounds with a `.json.example` extension in the same directory. These files do not load automatically.

If you want to use one of the examples, copy the file, remove the `.example` extension, replace its values with your own settings, and then put the file in the same directory or import it with the **From File** option.

---

## Address replacement and validation

During an Xray scan, BGScan builds a config in memory for each IP and replaces the destination address of the outbound with the target IP. The port stays the same and no temporary file is written to disk.

When adding an outbound, BGScan checks that it has a replaceable address and that the built config passes the internal Xray core validation.

Saved outbounds are checked again when the program starts. If an outbound is invalid, its error is reported during the initial program checks.

---

## Renaming and deleting an outbound

**Rename (`r`):** Select the outbound, press `r`, and enter the new name. The new name must not be a duplicate.

**Delete (`x`):** Select the outbound and press `x`. The file is removed permanently from `assets/xray/outbounds/` and there is no way to bring it back from inside the program.

---

## Related topics

- [Scan types](./scan-types.md)
- [Scan source](./scan-source.md)
- [Result files](./result-files.md)
