package keys

// Keys holds every binding in the app. It is built once at startup and
// passed down through ui.Env as env.Keys.
type Keys struct {
	Global  GlobalKeys
	Nav     NavKeys
	Table   TableKeys
	Crud    CrudKeys
	Filter  FilterKeys
	Confirm ConfirmKeys
	Notice  NoticeKeys
	Dialog  DialogKeys
	Menu    MenuKeys
	Form    FormKeys
	Scan    ScanKeys
	LogView LogViewKeys

	Inspector InspectorKeys

	Tab              TabKeys
	TextInput        TextInputKeys
	TextArea         TextAreaKeys
	SelectInput      SelectInputKeys
	MultiSelectInput MultiSelectInputKeys
	ToggleInput      ToggleInputKeys
	Picker           PickerKeys
}

// GlobalKeys holds the bindings that work in every mode.
type GlobalKeys struct {
	Quit, ForceQuit, DumpGoroutines, Help, Back, BackFromInput Binding
}

// NavKeys holds the shared movement bindings.
type NavKeys struct {
	Up, Down, Left, Right Binding
	PageUp, PageDown      Binding
	Top, Bottom           Binding
	NextTab, PrevTab      Binding
}

// ConfirmKeys holds the bindings for confirm prompts.
type ConfirmKeys struct {
	Accept Binding
}

// NoticeKeys holds the bindings for notice dialogs.
type NoticeKeys struct {
	Continue Binding
}

// DialogKeys holds the bindings for dialog overlays.
type DialogKeys struct {
	Left, Right Binding
	Cancel      Binding
}

// FormKeys holds the bindings for forms.
type FormKeys struct {
	Save   Binding
	Cancel Binding
}

// MenuKeys holds the bindings for menus.
type MenuKeys struct {
	Up, Down Binding
	Select   Binding
}

// ScanKeys holds the bindings for a running scan.
type ScanKeys struct {
	TogglePause, ToggleLogs, Cancel Binding
}

// LogViewKeys holds the bindings for the log view.
type LogViewKeys struct {
	Up, Down                 Binding
	PageUp, PageDown         Binding
	HalfPageUp, HalfPageDown Binding
	Top, Bottom              Binding
}

// InspectorKeys holds the bindings for the inspector panel.
type InspectorKeys struct {
	Up, Down         Binding
	NextTab, PrevTab Binding
	Edit             Binding
	Description      Binding
}

// TableKeys holds the bindings for result tables.
type TableKeys struct {
	Up, Down                        Binding
	PageUp, PageDown                Binding
	HalfPageUp, HalfPageDown        Binding
	Top, Bottom                     Binding
	ChangeSort, ToggleOrder, Filter Binding
}

// CrudKeys holds the bindings for list editing actions.
type CrudKeys struct {
	Select, Add, Delete, Rename Binding
}

// FilterKeys holds the bindings for table filter input.
type FilterKeys struct {
	Backspace Binding
	Clear     Binding
	Accept    Binding
}

// TextInputKeys holds the bindings for single line text input.
type TextInputKeys struct {
	Submit                  Binding
	CharacterForward        Binding
	CharacterBackward       Binding
	WordForward             Binding
	WordBackward            Binding
	DeleteWordBackward      Binding
	DeleteWordForward       Binding
	DeleteAfterCursor       Binding
	DeleteBeforeCursor      Binding
	DeleteCharacterBackward Binding
	DeleteCharacterForward  Binding
	LineStart               Binding
	LineEnd                 Binding
	Paste                   Binding
}

// TextAreaKeys holds the bindings for multi line text areas.
type TextAreaKeys struct {
	Submit                     Binding
	NewLine                    Binding
	CharacterForward           Binding
	CharacterBackward          Binding
	WordForward                Binding
	WordBackward               Binding
	LineNext                   Binding
	LinePrevious               Binding
	DeleteWordBackward         Binding
	DeleteWordForward          Binding
	DeleteAfterCursor          Binding
	DeleteBeforeCursor         Binding
	InsertNewline              Binding
	DeleteCharacterBackward    Binding
	DeleteCharacterForward     Binding
	LineStart                  Binding
	LineEnd                    Binding
	PageUp                     Binding
	PageDown                   Binding
	Paste                      Binding
	InputBegin                 Binding
	InputEnd                   Binding
	CapitalizeWordForward      Binding
	LowercaseWordForward       Binding
	UppercaseWordForward       Binding
	TransposeCharacterBackward Binding
	SelectCharacterForward     Binding
	SelectCharacterBackward    Binding
	SelectWordForward          Binding
	SelectWordBackward         Binding
	SelectLineUp               Binding
	SelectLineDown             Binding
	SelectAll                  Binding
	CopySelection              Binding
}

// SelectInputKeys mirrors huh's SelectKeyMap.
type SelectInputKeys struct {
	Submit       Binding
	Prev         Binding
	Next         Binding
	Up           Binding
	Down         Binding
	HalfPageUp   Binding
	HalfPageDown Binding
	GotoTop      Binding
	GotoBottom   Binding
	Left         Binding
	Right        Binding
	Filter       Binding
	SetFilter    Binding
	ClearFilter  Binding
}

// MultiSelectInputKeys mirrors huh's MultiSelectKeyMap.
type MultiSelectInputKeys struct {
	Submit       Binding
	Prev         Binding
	Next         Binding
	Toggle       Binding
	Up           Binding
	Down         Binding
	HalfPageUp   Binding
	HalfPageDown Binding
	GotoTop      Binding
	GotoBottom   Binding
	Filter       Binding
	SetFilter    Binding
	ClearFilter  Binding
	SelectAll    Binding
	SelectNone   Binding
}

// ToggleInputKeys holds the bindings for toggle inputs.
type ToggleInputKeys struct {
	Prev   Binding
	Next   Binding
	Submit Binding
	Toggle Binding
	Accept Binding
	Reject Binding
}

// PickerKeys mirrors the file picker's keymap.
type PickerKeys struct {
	Up       Binding
	Down     Binding
	PageUp   Binding
	PageDown Binding
	GotoTop  Binding
	GotoLast Binding
	Back     Binding
	Open     Binding
	Select   Binding
}

// TabKeys holds the bindings for moving between tabs.
type TabKeys struct {
	Next Binding
	Prev Binding
}
