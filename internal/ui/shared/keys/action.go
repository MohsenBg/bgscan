package keys

type Action string

// Global
const (
	ActionQuitNormal Action = "global.quit_normal"
	ActionQuitInput  Action = "global.quit_input"
	ActionHelp       Action = "global.help"
	ActionBack       Action = "global.back"
)

// Navigation
const (
	ActionUp       Action = "nav.up"
	ActionDown     Action = "nav.down"
	ActionLeft     Action = "nav.left"
	ActionRight    Action = "nav.right"
	ActionPageUp   Action = "nav.page_up"
	ActionPageDown Action = "nav.page_down"
	ActionTop      Action = "nav.top"
	ActionBottom   Action = "nav.bottom"
	ActionNextTab  Action = "nav.next_tab"
	ActionPrevTab  Action = "nav.prev_tab"
)

// Confirm
const (
	ActionAccept Action = "confirm.accept"
)

// Notice
const (
	ActionContinue Action = "notice.continue"
)

// Dialog
const (
	ActionDialogCancel Action = "dialog.cancel"
)

// Menu
const (
	ActionMenuSelect Action = "menu.select"
)

// Form
const (
	ActionFormSave Action = "form.save"
)

// Scan
const (
	ActionScanPause  Action = "scan.pause"
	ActionScanLogs   Action = "scan.logs"
	ActionScanCancel Action = "scan.cancel"
)
