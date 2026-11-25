// Package table provides a data table component for the terminal UI.
// It wraps the bubbles table model and adds sorting, text filtering, a
// loading state, and an empty state. A mutex guards the model, so
// callers can update rows from other goroutines while the UI reads them.
package table
