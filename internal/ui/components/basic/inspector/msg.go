package inspector

// tabChangeMsg is emitted by the tabs component when the active group
// changes, carrying the fields belonging to the newly selected group.
type tabChangeMsg struct {
	Group  string
	Fields []Field
}
