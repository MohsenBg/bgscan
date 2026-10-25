---
title: "Overview"
weight: 4
bookFlatSection: false
---

# Getting started with bgscan

This section introduces the program environment and shows you how to run your first scan.

## Running bgscan

Go into the directory where you installed the program and run it:

- **Windows:** Double click `bgscan.exe` or run `.\bgscan.exe` in PowerShell.
- **Linux/macOS/Termux:** Run `./bgscan` in the terminal.

## Main menu

When you run bgscan, the Splash screen appears first and then the program runs a few initial checks. At this stage it checks the settings, starts the internal Xray core, checks the optional Slipstream library, and also checks the settings for the DNS tunnels.

When the checks finish, press `Enter` to enter the program. Only an error reading the settings stops the program from running.

The main menu has these options:

| Option | Description |
|--------|-------------|
| **Run Scan** | Start a network scan |
| **IP Files** | Choose and manage the IP files for scanning |
| **Result Files** | View and manage saved results |
| **Xray Outbound** | Manage the Xray proxy settings |
| **DNS Tunneling** | Set up and manage DNS tunnel connections |
| **Settings** | Set the scan parameters and other preferences |
| **Logs** | View the program and scan logs |

## Navigation keys

Use these keys to move through the menus:

- **Arrow keys (↑ ↓):** Move between options
- **Enter:** Select the highlighted option
- **b or Esc:** Go back to the previous screen
- **q:** Quit the main menu with confirmation
- **Ctrl+C:** Quit the program immediately

> **Note:** On most screens you can press `?` to see the shortcuts for that screen.

## First scan

To run your first scan, follow these steps:

1. Run bgscan from the installation directory.
2. Choose **Run Scan** and press `Enter`.
3. Choose an IP list. To start, you can use one of the built-in lists such as `cloudflare_IPv4`.
4. Choose the scan type. For a simple connection test, `tcp` is a good option.
5. Press `Enter` to start the scan.
6. Wait for the scan to finish.
7. See the results there, or later enter **Result Files** from the main menu.

## What to read next

- [Scan types](../scanner/scan-types/): what each probe checks.
- [Scan pipeline](../scanner/scan-pipeline/): how targets are filtered before the scan starts.
- [IP lists](../scanner/ip-files/): how to import your own IPs.
- [Settings](../settings/overview/): set the workers, timeouts, and output.
- [User interface settings](../settings/ui/): set the theme, animation, table row count, and log level.
