package keys

type Action string

// Global
const (
	ActionQuit           Action = "global.quit_normal"
	ActionForceQuit      Action = "global.quit_force"
	ActionDumpGoroutines Action = "global.dump_goroutines"
	ActionHelp           Action = "global.help"
	ActionBack           Action = "global.back"
	ActionBackFromInput  Action = "global.back_from_input"
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

// Table
const (
	ActionTableHalfPageUp   Action = "table.half_page_up"
	ActionTableHalfPageDown Action = "table.half_page_down"
	ActionTableSort         Action = "table.sort"
	ActionTableOrder        Action = "table.order"
	ActionTableFilter       Action = "table.filter"
)

// CRUD
const (
	ActionCrudSelect Action = "crud.select"
	ActionCrudAdd    Action = "crud.add"
	ActionCrudDelete Action = "crud.delete"
	ActionCrudRename Action = "crud.rename"
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
