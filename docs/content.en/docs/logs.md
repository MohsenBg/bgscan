---
title: "Logs"
weight: 4
---

# Logs

bgscan maintains three separate log streams, each covering a different layer of the application. You can view any of them live from inside the TUI.

Navigate to **Main Menu → Logs** to open the log menu.

---

## Log Categories

| Key | Category | Description |
|---|---|---|
| `c` | Core Logs | Scan engine activity — IP probing, file I/O, result writing, and internal errors from the scanner core |
| `u` | UI Logs | Interface-layer events — component lifecycle, file operations triggered from the TUI, and UI-level errors |
| `d` | Debug Logs | Low-level diagnostic output — detailed internal state, data dumps, and verbose error traces for troubleshooting |

Press the corresponding key to open that log's viewer.

---

## The Log Viewer

Each log viewer shows a scrollable list of timestamped entries for that category. New messages stream in automatically while the viewer is open — you do not need to refresh manually.

Each line is formatted as:

```
2006/01/02 15:04:05 [LEVEL] message
```

Log levels are `INFO`, `WARN`, `ERROR`, and `DEBUG`.

The viewer pre-loads the last 200 lines from the log file when it opens, so you can see recent activity immediately even if nothing new is happening.

#### Navigation

| Key | Action |
|---|---|
| `↑` / `↓` | Scroll up / down |
| `b` or `Esc` | Close the log viewer |

---

## Log Files

All logs are written to the `logs/` directory next to the bgscan binary:

```
<bgscan-root>/
└── logs/
    ├── core.log
    ├── ui.log
    └── debug.log
```

Log files are rotated automatically — each file is capped at **50 MB**, up to **3 backup files** are kept, and files older than **7 days** are removed. Older backups are compressed to save space.

You can read these files directly with any text editor or `tail -f` if you need to inspect logs outside the TUI.
