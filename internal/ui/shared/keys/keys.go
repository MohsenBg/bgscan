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
