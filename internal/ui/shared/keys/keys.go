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

	TextInput        TextInputKeys
	TextArea         TextAreaKeys
	SelectInput      SelectInputKeys
	MultiSelectInput MultiSelectInputKeys
	ToggleInput      ToggleInputKeys
}

type GlobalKeys struct {
	Quit, ForceQuit, DumpGoroutines, Help, Back, BackFromInput Binding
}

type NavKeys struct {
	Up, Down, Left, Right Binding
	PageUp, PageDown      Binding
	Top, Bottom           Binding
	NextTab, PrevTab      Binding
}

type ConfirmKeys struct {
	Accept Binding
}

type NoticeKeys struct {
	Left, Right Binding
	Continue    Binding
}

type DialogKeys struct {
	Left, Right Binding
	Cancel      Binding
}

type FormKeys struct {
	Up, Down Binding
	Save     Binding
}

type MenuKeys struct {
	Up, Down Binding
	Select   Binding
}

type ScanKeys struct {
	TogglePause, ToggleLogs, Cancel Binding
}

type TableKeys struct {
	Up, Down                        Binding
	PageUp, PageDown                Binding
	HalfPageUp, HalfPageDown        Binding
	Top, Bottom                     Binding
	ChangeSort, ToggleOrder, Filter Binding
}

type CrudKeys struct {
	Select, Add, Delete, Rename Binding
}

type FilterKeys struct {
	Backspace Binding
	Clear     Binding
	Accept    Binding
}

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

type ToggleInputKeys struct {
	Prev   Binding
	Next   Binding
	Submit Binding
	Toggle Binding
	Accept Binding
	Reject Binding
}
