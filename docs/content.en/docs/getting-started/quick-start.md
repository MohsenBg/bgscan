---
title: "Overview"
weight: 4
bookFlatSection: false
---

# Getting Started with bgscan

What you see after installation, and how to run a first scan.

## Launching bgscan

Navigate to the installation folder and run the application:

- **Windows:** Double-click `bgscan.exe` or run `.\bgscan.exe` in PowerShell
- **Linux/macOS/Termux:** Run `./bgscan` in your terminal

## Main Menu Overview

When you launch bgscan, you'll see the main menu with the following options:

| Option | Description |
|--------|-------------|
| **Run Scan** | Start a new network scan |
| **IP Files** | Manage and select IP list files for scanning |
| **Result Files** | View and manage saved scan results |
| **Xray Outbounds** | Manage Xray proxy configurations |
| **DNS Tunneling** | Configure and manage DNS tunnel connections |
| **Settings** | Configure scan parameters and preferences |
| **Logs** | View application and scan logs |

## Navigation Controls

Use these keyboard shortcuts to navigate:

- **Arrow Keys** (↑ ↓) — Move up and down between menu items
- **Enter** — Select the highlighted option
- **b** or **Esc** — Go back to the previous screen
- **q** — Quit the application

> 💡 **Tip:** Press `q` at any time to exit bgscan safely.

## Running Your First Scan

1. **Launch bgscan** from your installation folder
2. Select **Run Scan** and press Enter
3. Choose an IP list (start with a built-in list like `cloudflare_IPv4`)
4. Select a scan type (try `tcp` for a quick connectivity test)
5. Press Enter to start the scan
6. Wait for the scan to complete
7. Review the results on the screen, or open Result Files from the main menu to view them later.

## What's Next?

- [Scan Types](../scanner/scan-types/) — what each probe measures
- [Scan Pipeline](../scanner/scan-pipeline/) — chaining stages so each one filters the next
- [IP Lists](../scanner/ip-files/) — importing your own targets
- [Settings](../settings/overview/) — tuning workers, timeouts, and output
