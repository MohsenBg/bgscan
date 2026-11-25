// Package http provides the settings inspector for the HTTP section of
// the scan configuration. It builds one validated input per config field
// on top of the shared inspector component. The accepted status codes
// are picked in two steps: whole code groups first, then single codes
// inside each selected group.
package http
