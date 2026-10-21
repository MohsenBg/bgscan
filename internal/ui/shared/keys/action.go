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
	ActionFormSave   Action = "form.save"
	ActionFormCancel Action = "form.cancel"
)

// Scan
const (
	ActionScanPause  Action = "scan.pause"
	ActionScanLogs   Action = "scan.logs"
	ActionScanCancel Action = "scan.cancel"
)

// Inspector
const (
	ActionInspectorEdit        Action = "inspector.edit"
	ActionInspectorDescription Action = "inspector.description"
)

// Tab
const (
	ActionTabNext     Action = "tab.next"
	ActionTabPrevious Action = "tab.previous"
)

// Text input
const (
	ActionTextInputSubmit                  Action = "textinput.submit"
	ActionTextInputCharacterForward        Action = "textinput.character_forward"
	ActionTextInputCharacterBackward       Action = "textinput.character_backward"
	ActionTextInputWordForward             Action = "textinput.word_forward"
	ActionTextInputWordBackward            Action = "textinput.word_backward"
	ActionTextInputDeleteWordBackward      Action = "textinput.delete_word_backward"
	ActionTextInputDeleteWordForward       Action = "textinput.delete_word_forward"
	ActionTextInputDeleteAfterCursor       Action = "textinput.delete_after_cursor"
	ActionTextInputDeleteBeforeCursor      Action = "textinput.delete_before_cursor"
	ActionTextInputDeleteCharacterBackward Action = "textinput.delete_character_backward"
	ActionTextInputDeleteCharacterForward  Action = "textinput.delete_character_forward"
	ActionTextInputLineStart               Action = "textinput.line_start"
	ActionTextInputLineEnd                 Action = "textinput.line_end"
	ActionTextInputPaste                   Action = "textinput.paste"
)

// Toggle input
const (
	ActionToggleInputPrev   Action = "toggleinput.prev"
	ActionToggleInputNext   Action = "toggleinput.next"
	ActionToggleInputSubmit Action = "toggleinput.submit"
	ActionToggleInputToggle Action = "toggleinput.toggle"
	ActionToggleInputAccept Action = "toggleinput.accept"
	ActionToggleInputReject Action = "toggleinput.reject"
)

// Picker
const (
	ActionPickerGotoTop  Action = "picker.goto_top"
	ActionPickerGotoLast Action = "picker.goto_last"
	ActionPickerUp       Action = "picker.up"
	ActionPickerDown     Action = "picker.down"
	ActionPickerPageUp   Action = "picker.page_up"
	ActionPickerPageDown Action = "picker.page_down"
	ActionPickerBack     Action = "picker.back"
	ActionPickerOpen     Action = "picker.open"
	ActionPickerSelect   Action = "picker.select"
)

// Select input
const (
	ActionSelectInputSubmit       Action = "selectinput.submit"
	ActionSelectInputPrev         Action = "selectinput.prev"
	ActionSelectInputNext         Action = "selectinput.next"
	ActionSelectInputUp           Action = "selectinput.up"
	ActionSelectInputDown         Action = "selectinput.down"
	ActionSelectInputHalfPageUp   Action = "selectinput.half_page_up"
	ActionSelectInputHalfPageDown Action = "selectinput.half_page_down"
	ActionSelectInputGotoTop      Action = "selectinput.goto_top"
	ActionSelectInputGotoBottom   Action = "selectinput.goto_bottom"
	ActionSelectInputLeft         Action = "selectinput.left"
	ActionSelectInputRight        Action = "selectinput.right"
	ActionSelectInputFilter       Action = "selectinput.filter"
	ActionSelectInputSetFilter    Action = "selectinput.set_filter"
	ActionSelectInputClearFilter  Action = "selectinput.clear_filter"
)

// Multi select input
const (
	ActionMultiSelectSubmit       Action = "multiselect.submit"
	ActionMultiSelectPrev         Action = "multiselect.prev"
	ActionMultiSelectNext         Action = "multiselect.next"
	ActionMultiSelectToggle       Action = "multiselect.toggle"
	ActionMultiSelectUp           Action = "multiselect.up"
	ActionMultiSelectDown         Action = "multiselect.down"
	ActionMultiSelectHalfPageUp   Action = "multiselect.half_page_up"
	ActionMultiSelectHalfPageDown Action = "multiselect.half_page_down"
	ActionMultiSelectGotoTop      Action = "multiselect.goto_top"
	ActionMultiSelectGotoBottom   Action = "multiselect.goto_bottom"
	ActionMultiSelectFilter       Action = "multiselect.filter"
	ActionMultiSelectSetFilter    Action = "multiselect.set_filter"
	ActionMultiSelectClearFilter  Action = "multiselect.clear_filter"
	ActionMultiSelectSelectAll    Action = "multiselect.select_all"
	ActionMultiSelectSelectNone   Action = "multiselect.select_none"
)

// Text area
const (
	ActionTextAreaSubmit                     Action = "textarea.submit"
	ActionTextAreaNewLine                    Action = "textarea.newline"
	ActionTextAreaCharacterForward           Action = "textarea.character_forward"
	ActionTextAreaCharacterBackward          Action = "textarea.character_backward"
	ActionTextAreaWordForward                Action = "textarea.word_forward"
	ActionTextAreaWordBackward               Action = "textarea.word_backward"
	ActionTextAreaLineNext                   Action = "textarea.line_next"
	ActionTextAreaLinePrevious               Action = "textarea.line_previous"
	ActionTextAreaDeleteWordBackward         Action = "textarea.delete_word_backward"
	ActionTextAreaDeleteWordForward          Action = "textarea.delete_word_forward"
	ActionTextAreaDeleteAfterCursor          Action = "textarea.delete_after_cursor"
	ActionTextAreaDeleteBeforeCursor         Action = "textarea.delete_before_cursor"
	ActionTextAreaInsertNewline              Action = "textarea.insert_newline"
	ActionTextAreaDeleteCharacterBackward    Action = "textarea.delete_character_backward"
	ActionTextAreaDeleteCharacterForward     Action = "textarea.delete_character_forward"
	ActionTextAreaLineStart                  Action = "textarea.line_start"
	ActionTextAreaLineEnd                    Action = "textarea.line_end"
	ActionTextAreaPageUp                     Action = "textarea.page_up"
	ActionTextAreaPageDown                   Action = "textarea.page_down"
	ActionTextAreaPaste                      Action = "textarea.paste"
	ActionTextAreaInputBegin                 Action = "textarea.input_begin"
	ActionTextAreaInputEnd                   Action = "textarea.input_end"
	ActionTextAreaCapitalizeWordForward      Action = "textarea.capitalize_word_forward"
	ActionTextAreaLowercaseWordForward       Action = "textarea.lowercase_word_forward"
	ActionTextAreaUppercaseWordForward       Action = "textarea.uppercase_word_forward"
	ActionTextAreaTransposeCharacterBackward Action = "textarea.transpose_character_backward"
	ActionTextAreaSelectCharacterForward     Action = "textarea.select_character_forward"
	ActionTextAreaSelectCharacterBackward    Action = "textarea.select_character_backward"
	ActionTextAreaSelectWordForward          Action = "textarea.select_word_forward"
	ActionTextAreaSelectWordBackward         Action = "textarea.select_word_backward"
	ActionTextAreaSelectLineUp               Action = "textarea.select_line_up"
	ActionTextAreaSelectLineDown             Action = "textarea.select_line_down"
	ActionTextAreaSelectAll                  Action = "textarea.select_all"
	ActionTextAreaCopySelection              Action = "textarea.copy_selection"
)
