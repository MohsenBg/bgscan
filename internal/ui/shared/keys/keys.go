package keys

// Keys holds every binding in the app. It is built once at startup and
// passed down through ui.Deps as deps.Keys.
type Keys struct {
	Global  GlobalKeys
	Nav     NavKeys
	Confirm ConfirmKeys
	Notice  NoticeKeys
	Dialog  DialogKeys
	Menu    MenuKeys
	Form    FormKeys
	Scan    ScanKeys
}

type GlobalKeys struct {
	Quit, Help, Back, BackInputMode Binding
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
	Continue Binding
}

type DialogKeys struct {
	Cancel Binding
}

type MenuKeys struct {
	Select Binding
}

type FormKeys struct {
	Save Binding
}

type ScanKeys struct {
	Pause, Logs, Cancel Binding
}
