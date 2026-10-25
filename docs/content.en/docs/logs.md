---
title: "Logs"
weight: 5
---

# Logs

BGScan records two kinds of logs. One belongs to the scanner core and the other belongs to the terminal interface. You can view both live from inside the TUI.

To reach the logs, go to **Main Menu → Logs**.

## Log types

| Key | Section | Description |
|---|---|---|
| `c` | Core Logs | The scanner core activity, such as running probes on IPs, file operations, saving results, and internal errors |
| `u` | UI Logs | Interface events, such as the component lifecycle, file operations started from inside the TUI, and interface errors |

Press the key for each section to see its logs.

## Log level

Both log types use the `log_level` value in `settings/ui_settings.toml`:

```toml id="8v4m2k"
log_level = "info"
```

The allowed values are `debug`, `info`, `warn`, and `error`.

This value is applied to both loggers when the program starts. If you change it from **Settings → UI Settings**, the new setting takes effect right away and BGScan does not need to restart.

The `debug` level records more details, including probe errors with the full target IP. This level is useful for troubleshooting, but it also increases the log size. The default level is `info`.

## Viewing logs

Each viewer shows a scrollable list of timestamped messages for that section. New messages are added automatically while the viewer is open, so there is no need to refresh by hand.

Each log line has this format:

```text
2006/01/02 15:04:05 [LEVEL] message
```

The message levels are `INFO`, `WARN`, `ERROR`, and `DEBUG`. Which messages appear depends on the `log_level` value.

When a viewer opens, the last 200 lines of the log file are read so you can see recent activity right away, even if nothing new has happened.

### Movement keys

| Key | Action |
|---|---|
| `↑` / `↓` | Move up and down the list |
| `b` or `Esc` | Close the viewer |

## Where log files are stored

Log files are stored in the `logs/` directory next to the BGScan binary:

```text
<bgscan-root>/
└── logs/
    ├── core.log
    └── ui.log
```

Log files rotate automatically so they do not take up much space. The limits are:

- Maximum size of each file: **50 megabytes**
- Number of backup files: at most **3 files**
- Removal of files older than **7 days**
- Compression of older backup files to reduce space use

If you want to check the logs outside the TUI, you can open the files with any text editor or use the `tail -f` command to watch messages live.
