package inspector

// tabChangeMsg is emitted when the active group changes.
type tabChangeMsg struct {
	Group  string
	Fields []Field
}
