<div align="right">

[**English**](./xray_outbounds.md)  |  [**فارسی**](./xray_outbounds.fa.md)

</div>

# Adding a Custom Xray Outbound

## Table of Contents

- [Method 1: Interactive Outbound Wizard](#method-1-interactive-outbound-wizard)
  - [Option A: Add via Link](#option-a-add-via-link)
  - [Option B: Add via JSON File](#option-b-add-via-json-file)
  - [Result](#result)
- [Method 2: Manual Template Editing (Advanced)](#method-2-manual-template-editing-advanced)
  - [1. Navigate to the Outbounds Directory](#1-navigate-to-the-outbounds-directory)
  - [2. Copy and Rename a Template](#2-copy-and-rename-a-template)
  - [3. Edit the Configuration File](#3-edit-the-configuration-file)
  - [4. Using Your Outbound in BGScan](#4-using-your-outbound-in-bgscan)
---

BGScan supports two ways to add a custom Xray outbound:

1. **Interactive Outbound Wizard** (recommended) — add an outbound directly from the app using a share link (VLESS / VMess / Trojan / Shadowsocks) or a JSON file.
2. **Manual Template Editing** — copy a template file and edit it by hand.

---

## Method 1: Interactive Outbound Wizard

This is the fastest way to add an outbound — no manual file editing required.

### Steps

1. Open **BGScan**.
2. Navigate to:
   ```
   Xray → Outbounds
   ```
3. Press the **a** key/button to open the **Add Outbound** dialog.

![menu](https://github.com/user-attachments/assets/d2a38b21-bf09-44b7-bca4-4267a284740b)

4. In the dialog, choose how you want to add the outbound:

   - **From Link** — VLESS, VMess, Trojan, or Shadowsocks share link
   - **From JSON File** — an Xray-compatible outbound JSON file

---

### Option A: Add via Link

1. Select **From Link**.

![add_via_link](https://github.com/user-attachments/assets/eb23f4de-945d-4739-a504-176d8b73edf7)

2. Paste your share link when prompted (e.g. `vless://...`, `vmess://...`, `trojan://...`, `ss://...`).

![paste link](https://github.com/user-attachments/assets/c5683fd3-b777-46b4-ba24-3f9938db5b78)

3. BGScan automatically parses the link and extracts the protocol, server, port, credentials, and transport settings.
4. Enter a **name** for the outbound when prompted.

![enter name](https://github.com/user-attachments/assets/5ecea6ee-6602-4671-8a4d-6b5cd2510029)

5. The outbound is saved and immediately available in the **Outbounds** list.

```
Add Outbound → From Link → Paste Link → Enter Name → Done
```

> Supported link formats: `vless://`, `vmess://`, `trojan://`, `ss://`

---

### Option B: Add via JSON File

1. Select **From JSON File**.

![add via json](https://github.com/user-attachments/assets/2ef7ad7c-52b9-4603-b757-cd2bd3e826f3)

2. Browse and select the outbound `.json` file you want to import.

![select json file](https://github.com/user-attachments/assets/a5cd36d3-bdfa-4ed0-a348-0a7962336e56)

3. Enter a **name** for the outbound when prompted.
4. The outbound is saved and immediately available in the **Outbounds** list.

```
Add Outbound → From JSON File → Select File → Enter Name → Done
```

> ⚠️ **JSON file format requirement:**
> The selected JSON file must contain a **single outbound object**, using the exact same format as the manual template (see [Method 2](#method-2-manual-template-editing-advanced)) — including the `"address": "$ADDRESS"` placeholder, which BGScan replaces automatically during testing. Do not provide a full Xray config (with `outbounds: [...]`, `inbounds`, `routing`, etc.) — only the single outbound block itself.
>
> Minimal valid example:
> ```json
> {
>   "protocol": "vless",
>   "settings": {
>     "vnext": [
>       {
>         "address": "$ADDRESS",
>         "port": 443,
>         "users": [
>           { "id": "your-uuid-here", "encryption": "none" }
>         ]
>       }
>     ]
>   },
>   "streamSettings": {
>     "network": "ws",
>     "security": "tls",
>     "tlsSettings": { "serverName": "example.com" },
>     "wsSettings": { "path": "/ws", "headers": { "Host": "example.com" } }
>   }
> }
> ```
> If the file doesn't match this format (e.g. missing `$ADDRESS`, or wrapped in a full config), the import will fail or the outbound won't be testable.

---

### Result

Once added (via either method), your new outbound appears in the list:

![outbounds](./images/outbounds.png)

You can re-run the wizard at any time to add more outbounds — no need to touch the filesystem.

---

## Method 2: Manual Template Editing (Advanced)

If you prefer full manual control over the configuration, you can edit a template file directly.

### 1. Navigate to the Outbounds Directory

All outbound templates are stored in:

```
assets/xray/outbounds/
```

Directory example:

```
assets/xray/outbounds/
├── vless_grpc.json.example
├── vless_ws.json.example
├── vless_ws_no_tls.json.example
├── vless_xhttp.json.example
└── vless_xhttp_no_tls.json.example
```

Files with the `.example` extension are templates.

---

### 2. Copy and Rename a Template

Choose a template (for example: `vless_ws.json.example`) and copy it:

```bash
cp vless_ws.json.example config.json
```

Updated directory:

```
assets/xray/outbounds/
├── vless_grpc.json.example
├── vless_ws.json.example
├── vless_ws_no_tls.json.example
├── vless_xhttp.json.example
├── vless_xhttp_no_tls.json.example
└── config.json
```

You can name the file anything you like (`my_ws.json`, `tls_ws.json`, etc.).

---

### 3. Edit the Configuration File

Open your file:

```
assets/xray/outbounds/config.json
```

Replace all fields marked with `?` using your real outbound values.

> ⚠️ **Important:** Do **not** edit the `address` field:
> ```json
> "address": "$ADDRESS"
> ```
> The scanner replaces `$ADDRESS` automatically during testing.

---

#### Example Template (Before Editing)

```json
{
  "tag": "proxy",
  "protocol": "vless",
  "settings": {
    "vnext": [
      {
        "address": "$ADDRESS",
        "port": 443,
        "users": [
          {
            "id": "?",
            "encryption": "none"
          }
        ]
      }
    ]
  },
  "streamSettings": {
    "network": "ws",
    "security": "tls",
    "tlsSettings": {
      "allowInsecure": false,
      "serverName": "?",
      "alpn": ["h2", "http/1.1"],
      "fingerprint": "firefox"
    },
    "wsSettings": {
      "path": "?",
      "headers": {
        "Host": "?"
      }
    }
  }
}
```

#### Example (After Filling Values)

```json
{
  "tag": "proxy",
  "protocol": "vless",
  "settings": {
    "vnext": [
      {
        "address": "$ADDRESS",
        "port": 443,
        "users": [
          {
            "id": "3f1e6f4c-9f1c-4a3a-bf10-9e2c8a123456",
            "encryption": "none"
          }
        ]
      }
    ]
  },
  "streamSettings": {
    "network": "ws",
    "security": "tls",
    "tlsSettings": {
      "allowInsecure": false,
      "serverName": "example.com",
      "alpn": ["h2", "http/1.1"],
      "fingerprint": "firefox"
    },
    "wsSettings": {
      "path": "/ws",
      "headers": {
        "Host": "example.com"
      }
    }
  }
}
```

You must fill in:

| Field | Description |
|---|---|
| `id` | Your VLESS UUID |
| `serverName` | TLS SNI |
| `path` | WebSocket path |
| `Host` | Host header |

---

### 4. Using Your Outbound in BGScan

After saving the JSON file:

1. Open **BGScan**.
2. Navigate to:
   ```
   Xray → Outbounds
   ```
3. Your new outbound will appear automatically in the list.

![menu](https://github.com/user-attachments/assets/583d4d51-a718-40c0-9272-292c4d0c9c1b)
![outbounds](https://github.com/user-attachments/assets/99c9e713-2811-4787-8247-4056dfeefb81)

