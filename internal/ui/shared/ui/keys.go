package ui

type Action string

type Binding struct {
	Action Action   // stable name: "move_left", "select"
	Keys   []string // what the code matches: ["left", "h"]

	// Help display
	HelpKey   string // compact key label: "←" (defaults to Keys[0])
	HelpText  string // precise label for long help: "move left"
	ShortText string // optional label for short help; empty = use HelpText

	Hidden   bool // works, but omit from short help
	Disabled bool // not active right now, omit everywhere
}
