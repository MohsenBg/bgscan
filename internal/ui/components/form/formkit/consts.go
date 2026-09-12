package formkit

const (
	// MaxFormWidth is the default maximum form body width.
	MaxFormWidth = 55
	// MaxFormHeight is the default maximum form body height.
	MaxFormHeight = 40
	// FormPadding is the horizontal padding around the form body.
	FormPadding = 2
)

const (
	// GroupConnection groups base tunnel/connection fields.
	GroupConnection = "Connection"
	// GroupAdvanced groups tunable/performance fields.
	GroupAdvanced = "Advanced"
	// GroupProxyAuth groups proxy and authentication fields.
	GroupProxyAuth = "Proxy & Auth"
)

// AlwaysVisible is the Visible predicate for fields that are always shown.
func AlwaysVisible() bool { return true }
