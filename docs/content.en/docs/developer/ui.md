---
title: "User interface"
weight: 4
---

# User interface

The bgscan terminal interface is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) version 2 (`charm.land/bubbletea/v2`) and uses [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling. The interface is a tree of components, dialogs sit on top of the screen as overlays, and a central manager controls the layout.

---

## Package map

| Package | Responsibility |
|---|---|
| `ui/main/app` | The main Bubble Tea model that manages the three app stages: Splash, then Startup, then Workspace |
| `ui/main/splash` | The opening screen with the ASCII logo animation and the version number |
| `ui/main/startup` | The initial checks. It checks the loggers, registers the probe schemas, and validates Config, Xray, DNSTT, Slipstream, VayDNS, MasterDNS, StormDNS, TheFeed, and the app state |
| `ui/main/workspace` | The main program environment with the Body, the Footer, and the stack of overlay dialogs |
| `ui/main/body` | The main content area that holds the component stack, from the main menu to the different screens |
| `ui/main/footer` | The status bar with the app version, the current screen name, the Goroutine count, and memory use |
| `ui/components/basic` | Reusable widgets such as confirm, crud, form, input, inspector, logview, menu, notice, picker, progress, table, and tabs |
| `ui/components/form` | Settings forms for the DNSTT, Slipstream, VayDNS, MasterDNS, StormDNS, and TheFeed tunnels, built on a shared `formkit` |
| `ui/components/inspector` | Settings forms for the protocols and app sections: DNS, General and Writer, HTTP, ICMP, TCP, Xray, and UI |
| `ui/components/menus` | The main and secondary menus such as entry, logs, outbound, scantype, settings, targetsource, and dnstunmenu |
| `ui/components/tables` | Data tables including iplist, ipviewer, outbounds, resultlist, and dnstun for managing DNS tunnels |
| `ui/components/scanner` | The live scan view with a staged progress bar and a result table for each stage |
| `ui/shared` | Shared interface infrastructure: layout, dialogs, Session and Env, keys, the component interface, animation, Meter, and validation |
| `ui/theme` | 21 dark and light themes, automatic theme detection from the terminal background, and a theme adapter for huh forms |

---

## Component interface

Every screen, widget, and dialog implements the `ui.shared.ui.Component` interface:

```go
type Component interface {
    ID() ComponentID
    Name() string
    Init() tea.Cmd
    Update(tea.Msg) (Component, tea.Cmd)
    View() string
    OnClose() tea.Cmd
    Mode() KeyMode
}
```

- **`ID`**: A UUID built with `NewComponentID()` and used to identify the component in open and close messages.
- **`Name`**: The readable component name shown in the Footer status bar.
- **`Init`**: Runs once when the component is added to the interface. It returns a `tea.Cmd` for asynchronous work.
- **`Update`**: Processes incoming messages and returns the updated component along with the commands it needs.
- **`View`**: Builds the component's text output for display in the terminal.
- **`OnClose`**: Does cleanup work when the component is removed from the stack.
- **`Mode`**: Defines which keys are used for Back and Quit. The [input modes](#input-modes) section has the details.

---

## Main app model (`app`)

The `ui/main/app` package implements the main Bubble Tea model. This model manages the different stages of the program and the shared interface state:

```go
type model struct {
    session   *ui.Session
    splash    ui.Component
    startup   ui.Component
    workspace ui.Component
    stage     AppStage
    // dialog stack
}

func New(uiLogger, coreLogger logger.Logger) Application
```

The program goes through three stages:

| Stage | Component | Description |
|---|---|---|
| `StageSplash` | `splash` | The ASCII logo with a Glitch animation and the version number. Any key skips it. |
| `StageStartUP` | `startup` | The initial checks run in order: Logger, Config, Xray, DNSTT, Slipstream, VayDNS, MasterDNS, StormDNS, TheFeed, App |
| `StageWorkspace` | `workspace` | The main program environment with the Body, the Footer, and the stack of overlay dialogs |

`ui.Session` and `ui.Env` replace the earlier flat `AppState` struct. Session provides the shared dependencies to the different components:

```go
type Session struct {
    Config  *config.ScannerConfig
    Store   *config.Store
    Program ui.Program
    // env and core logger
}

type Env struct {
    Layout   layout.Layout
    Theme    theme.Theme
    Keys     *keys.Keys
    Log      logger.Logger // ui.log
    Settings UISettings
}
```

Components read settings from `session.Config` and save changes through `session.Store`. The theme, layout, keys, and interface settings are available from `session.Env()`. Config does not use a global singleton.

### Message routing

How messages are processed depends on the current app stage:

1. **Splash:** The animation runs until it finishes or the user presses a key, then the program moves to Startup.
2. **Startup:** Check messages arrive through the `logCh` channel from a Goroutine. Critical errors stop the following checks. Pressing Enter moves the program to Workspace.
3. **Workspace:** Interface messages are processed at this stage:

   - `tea.WindowSizeMsg`: On a terminal size change, `layout.Update(width, height)` runs again.
   - `tea.KeyPressMsg`: The `ctrl+c` key closes the program immediately, and `ctrl+t` writes Goroutine details to `ui.log` at the Debug level.
   - **Overlay management:** When a dialog is open, the topmost dialog receives keyboard input. The Back and Quit keys close it.
   - **Navigation at the root:** In the main menu, `b` or `q` open a quit confirmation dialog instead of closing immediately.
   - `dialog.OpenDialogMsg`: Adds a new overlay to the stack with its position details.
   - `ui.CloseComponentMsg`: Removes the overlay matching the ID from the stack and runs its `OnClose`.
   - **Sending general messages:** If no overlay receives a message, it is sent to the Body and the Footer.

### Display

In the **Workspace** stage, the app layout looks like this:

```text
┌─────────────────────────────────┐
│                                 │
│           body                  │  ← component stack (menu → screen)
│                                 │
├─────────────────────────────────┤
│           footer                │  ← version · status · goroutines · mem
└─────────────────────────────────┘
    + overlay dialogs (centered/positioned)
```

In version 2.12 there is no separate header and the Banner appears inside the main menu. Overlays sit on top of the main content using `bubbletea-overlay`.

The minimum acceptable terminal size is 75x35. If the size drops below that, the `terminal too small` message appears.

---

## Splash screen

The `ui/main/splash` package shows the animated ASCII logo "BGSCAN" with a Glitch effect:

- 80 animation frames with a 30 millisecond gap between ticks, then 150 fixed frames
- Some logo characters are replaced with Glitch characters such as `▓▒░╬╪┼╫┃━╳▄▀`, and the chance of replacement drops through the animation
- A horizontal highlight in the Accent color appears at 72 percent progress
- Horizontal shaking and broken image rows appear during the animation
- The version number fades in after the animation ends

---

## Startup checks

The `ui/main/startup` package runs ten checks in order inside one Goroutine. The status of each check appears in the side bar with Pending, Running, Success, Warn, and Error markers. The progress bar and the scrollable section also show the check details.

| # | Category | What is checked |
|---|---|---|
| 1 | Logger | The Core and UI loggers are checked and the built-in probe schemas are registered through `core.Init()` |
| 2 | Config | Settings are loaded, `NormalizeAll` runs, and the corrected values are reported |
| 3 | Xray | The internal core version is shown and the outbounds are listed and validated |
| 4 | DNSTT | All DNSTT settings files are validated |
| 5 | Slipstream | The client library is found and loaded, checked, and the Slipstream settings are validated |
| 6 | VayDNS | All VayDNS settings files are validated |
| 7 | MasterDNS | All MasterDNS settings files are validated |
| 8 | StormDNS | All StormDNS settings files are validated |
| 9 | TheFeed | All TheFeed settings files are validated |
| 10 | App | The program waits for the user to press Enter |

Critical errors stop the following checks. You can scroll through the output with the `j` and `k` keys, and once the checks finish you enter the Workspace by pressing Enter.

---

## Body and component stack

The `ui/main/body` package holds a stack of components. The component at the top of the stack is the active screen. Body is active only in the `workspace` stage and is not used during Splash and Startup.

```text
components[0] = mainMenu (entry)    ← always in the stack
components[1] = scan source picker  ← pushed onto the stack
components[2] = scan type picker    ← pushed onto the stack
components[3] = scanner view        ← pushed onto the stack
```

**Navigation messages**

| Message | Action |
|---|---|
| `ui.OpenComponentMsg` | Adds a new component to the top of the stack |
| `ui.CloseComponentMsg` | Removes the component identified by ID |
| `ui.ResetComponentStacksMsg` | Empties the stack back to the main menu |
| Back key (`b` / `Esc`) | Removes the top component, as long as the stack has more than one component |
| Quit key (`q`) | Opens the quit confirmation dialog |

Only the component at the top of the stack receives the `Update` message. When that component is removed, the next component becomes active and its name appears in the Footer status bar.

---

## Overlay dialog system

The `ui/shared/dialog` package makes it possible to show dialogs over the main layout.

```go
type OpenDialogMsg struct {
    Component ui.Component
    XPos      DialogPosition // Top, Right, Bottom, Left, Center
    YPos      DialogPosition
    XOffset   int
    YOffset   int
}
```

Dialog positions use the `rmhubbert/bubbletea-overlay` library. Most dialogs appear in the center of the screen, but the file picker and some confirmation dialogs may use a different position.

The dialog stack in `app.model` allows several overlays to be open on top of each other. The topmost dialog receives all input. Opening and closing dialogs is animated, and on close the `OnClose` call waits until the animation finishes.

If you turn off the **Animation** option in the interface settings, dialog animations and the other interface animations are disabled.

---

## Input modes

The `ui/shared/ui/mode.go` file defines three modes that control keyboard key behavior:

| Mode | Use | Back keys | Note |
|---|---|---|---|
| `KeyModeNormal` | Normal navigation | `b`, `Esc` | The `q` key opens the quit confirmation, and in the main menu `b` and `Esc` do the same |
| `KeyModeInput` | Typing in a text field or picker | `Esc` | When a field is in input mode, `q` and `b` are treated as text |
| `KeyModeManage` | Screens with dedicated keys, such as tables and the scanner | Depends on the component | The scanner uses `Esc`, `b`, or `q` to open the cancel confirmation |

Key definitions live in `ui/shared/keys/defaults.go` and are available to components through `session.Env().Keys`. Components use the defined Actions instead of comparing key strings directly. This keeps Back, Quit, navigation, table, and input key behavior the same across different sections.

The Footer and dialog hints use these same definitions. Pressing `?` shows the key list. The `Ctrl+C` key closes the program immediately at the app level, whatever the current mode is.

### Keyboard layout warning

If a non Latin character reaches Workspace in `KeyModeNormal`, bgscan shows a **Keyboard Layout** notice explaining that shortcuts need an English layout.

This warning does not appear in Input and Manage modes, while a dialog is open, or outside Workspace. Repeated notices do not pile up either.

---

## Layout

The `ui/shared/layout/layout.go` file calculates the size and position of the interface sections. Whenever the terminal size changes, the dimensions of the different areas are recalculated.

```go
type Layout struct {
    terminal TerminalSize // actual terminal dimensions
    content  ContentSize  // drawable space after margins and padding
    body     ComponentSize
    footer   ComponentSize
}
```

In version 2.12, Workspace has only two drawable sections: Body and Footer. The old header and the `ui/main/header` package are no longer used.

On every `tea.WindowSizeMsg`, the `Layout.Update(width, height)` method runs. Components get the shared layout from `session.Env().Layout`.

The default terminal size is 80x35 and the minimum acceptable size is 75x35.

---

## Reusable widgets (`basic`)

The `ui/components/basic/` directory contains widgets used across different screens:

| Widget | Use |
|---|---|
| `confirm` | A yes or no confirmation dialog for actions such as quitting and deleting |
| `crud` | A general CRUD interface for tables such as the IP list and the outbounds, built on the `Provider` abstraction |
| `input` | The input dialog shell and typed widgets such as textinput, textarea, selectinput, multiselect, and toggleinput |
| `inspector` | A general key and value editor for settings, with field formatting and type adapters |
| `logview` | A live scrollable view of the Core and UI logs |
| `menu` | A vertical menu controlled by the keyboard |
| `notice` | A temporary notice in a banner layout |
| `picker` | A file selection dialog |
| `progress` | An animated progress bar with ETA and rate display |
| `table` | A data table with sorting and scrolling |
| `tabs` | A tabbed container. The scanner view uses this widget to show each stage |

Each widget implements the standard `Component` interface and can be used in larger screens.

---

## Settings inspectors and forms

### Settings inspectors (`ui/components/inspector/`)

| Form | Settings file |
|---|---|
| `general` | `general_settings.toml` |
| `icmp` | `icmp_settings.toml` |
| `tcp` | `tcp_settings.toml` |
| `http` | `http_settings.toml` |
| `dns` | `dns_settings.toml` |
| `xray` | `xray_settings.toml` |
| `ui` | `ui_settings.toml` |

Each inspector reads its own section from `session.Config`, shows the editable fields, and calls the matching `session.Store.SaveXxx()` method when saving.

The UI settings inspector uses `session.Env().Settings` and saves changes with `ui.SaveSettings()`.

Fields can depend on earlier choices. For example, TLS settings appear only when HTTPS is selected, and tunnel specific fields appear only for the matching protocol.

Every field is validated before it is saved. The inspector applies changes to a copy of the settings section and runs `validate.ValidateXxx` on it. If a field is invalid, the change is rejected and the invalid value never enters Config or the settings file.

The general `basic/inspector` widget handles field display, formatting, and type conversion. Each protocol inspector only defines which fields to show and how they map to the settings struct.

### DNS tunnel settings forms (`ui/components/form/`)

| Form | Protocol | Settings struct |
|---|---|---|
| `dnstt` | DNSTT | `dns.DNSTTConfig` |
| `slipstream` | Slipstream | `dns.SlipstreamConfig` |
| `vaydns` | VayDNS | `dns.VayDNSConfig` |
| `masterdns` | MasterDNS | `dns.MasterDNSConfig` |
| `stormdns` | StormDNS | `dns.StormDNSConfig` |
| `thefeed` | TheFeed | `dns.TheFeedConfig` |

Each form builds the inspector for its own protocol and shows the connection fields. Advanced options exist only in the VayDNS form, and proxy and authentication settings are offered in the DNSTT, VayDNS, and Slipstream forms.

These forms are used both to create a new config and to edit an existing one. When saving, the matching service's `SaveConfig` or `EditConfig` method is called. The field builders are in `ui/components/form/formkit`.

---

## Menus and tables

### Menus (`ui/components/menus`)

| Menu | Where to reach it | Use |
|---|---|---|
| `entry` | Program start | The main menu with Run Scan, IP Files, Result Files, Xray Outbound, DNS Tunneling, Settings, and Logs |
| `targetsource` | Run Scan | Chooses the scan input from an IP list or earlier results |
| `scantype` | After choosing the input | Chooses the scan type: ICMP, TCP, HTTP, Xray, DNS Resolve, or DNS Tunneling |
| `settings` | The Settings option | Chooses the settings section |
| `outboundmenu` | The Xray Outbound option | Manages the outbounds |
| `dnstunmenu` | The DNS Tunneling scan type | Chooses the protocol from DNSTT, Slipstream, VayDNS, MasterDNS, StormDNS, and TheFeed |
| `logs` | The Logs option | Chooses the Core or UI log |

### Tables (`ui/components/tables`)

| Table | Data source | Provider |
|---|---|---|
| `iplist` | The `ips/` directory | `iplist.Registry` |
| `resultlist` | The `result/` directory | `result.GetResultFiles` |
| `outbounds` | The outbound settings in Xray | The `xray` package |
| `dnstun` | The DNS tunnel settings | `dns.GetAllDNSTunsFile` |
| `ipviewer` | The results held in memory | The scanner component |

Tables use the general `basic/table` widget and the `Provider` interface. The provider supplies the rows and handles column sorting.

The result browser first shows general file information. When you open a file, `ipviewer` builds the columns from the schema attached to that file, so results from registered probes are displayed without changing the table code.

The MasterDNS and StormDNS schemas are not registered in the registry, so the result files for these two protocols do not appear in the result browser. The `dnstun` table also lets you manage the configs of all six DNS tunnel protocols.

---

## Scanner view

The `ui/components/scanner` package is the live scan dashboard and connects the Core scanner to the user interface.

This section offers:

- **Stage tab bar:** Shows the status of each stage live with states such as `waiting`, `scanning`, `paused`, `done`, and `error`.
- **Progress bar:** Shows the number of processed IPs and the total, the estimated time remaining (ETA), the processing rate in IP/s, and a Sparkline of recent rates.
- **Scan and Errors panels:** The Scan panel shows the hit rate. The Errors panel groups errors into categories such as timeout, refused, unreachable, closed, badresponse, and other errors, and shows the count and percentage of each category.
- **Result table for each stage:** Uses `tables/ipviewer` and builds the columns from that stage's schema. Time and bandwidth values use different colors based on quality.
- **Independent stage state:** Each stage moves from `Waiting` to `PreProcessing`, then to `Scanning`, and finally to `Ended` or `Error`.
- **Live updates:** Progress details come from `engine.ScanHooks.OnProgress`, and successful results enter the table in batches.

The main keys while scanning are:

- `p` or `space`: pause and resume the scan
- `l`: open the Core log viewer
- `Tab`: switch between stages
- `Esc`, `b`, or `q`: open the scan cancel confirmation dialog

---

## Theme

The `ui/theme/theme.go` file defines the central color palette of the interface:

```go
type Theme struct {
    Primary       color.Color
    Secondary     color.Color
    Border        color.Color
    BorderActive  color.Color
    Text          color.Color
    Muted         color.Color
    Timestamp     color.Color
    Info          color.Color
    Error         color.Color
    Success       color.Color
    Orange        color.Color
    Yellow        color.Color
    Purple        color.Color
    ProgressStart color.Color
    ProgressEnd   color.Color
}
```

### Built-in themes

The `ui/theme` package has 21 built-in themes: 13 dark themes and 8 light themes. The theme name is read from the `theme` option in `settings/ui_settings.toml`.

If this option is empty, the program detects the terminal background through the `$COLORFGBG` environment variable and chooses the `bgscan-dark` or `bgscan-light` theme. If the theme name is unknown, the setting falls back to automatic selection.

The theme list:

`bgscan-dark`, `tokyo-night`, `gruvbox-dark`, `one-dark`, `dracula`, `monokai`, `solarized-dark`, `rose-pine`, `kanagawa`, `github-dark`, `material-dark`, `everforest-dark`, `ayu-dark`, `bgscan-light`, `gruvbox-light`, `one-light`, `solarized-light`, `rose-pine-dawn`, `everforest-light`, `ayu-light`, and `github-light`.

Components get colors from `session.Env().Theme` and should not hard code colors in their own code. The `Session.SetTheme` method saves the selected theme and sends a `theme.ThemeChanged` message so cached forms and widgets are rebuilt.

Each theme also defines the `TierExcellent`, `TierGood`, `TierFair`, and `TierPoor` colors. Failed or out of range values use that theme's Muted and Error colors. The IP Viewer uses these colors to show latency and bandwidth quality.

### huh integration

The `ui/theme/huh.go` file builds a `huh.Theme` from the central palette. This keeps form fields such as select, confirm, and text input, which the `huh` library draws, in line with the rest of the interface. The input widgets and the inspectors also use this theme.

---

## Troubleshooting

### Taking a Goroutine dump

Pressing `Ctrl+T` runs the `dumpGoroutines()` function in `app/update.go`, which records the following details:

- It writes the total Goroutine count to the log.
- It records the full Goroutine profile through `runtime/pprof` at the Debug level in `ui.log`.
- It separates and logs the stacks that contain `[chan receive]`, `[chan send]`, `[select]`, `[IO wait]`, or `[sleep]`. This information is useful for finding stuck workers or probes that have not released their resources.

Because this information is recorded at the `Debug` level, set the following value in `settings/ui_settings.toml` to see it:

```toml
log_level = "debug"
```

### Runtime statistics

The Footer component reads Go runtime statistics every second:

- `runtime.NumGoroutine()`: the Goroutine count
- `runtime.ReadMemStats()`: Heap and Sys memory details

These values appear in the status bar and are useful for checking resource use during a scan.

---

## Related pages

- [Architecture](./architecture/): an introduction to the overall project structure
- [Core](./core/): an introduction to the scan engine and the Probe interface
