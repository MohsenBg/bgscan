// Package ui defines the shared types that tie the TUI together.
// It has the Component interface every screen implements, the messages
// that open and close components, the Session and Env values that carry
// config, theme, keys and loggers to components, and the persisted
// UISettings. Components read shared state from Env and never build
// their own session.
package ui
