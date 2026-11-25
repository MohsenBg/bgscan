// Package theme defines the color palettes the TUI renders with.
// A Theme is a flat set of named colors. Built-in palettes are declared
// in default.go and enter the registry through Init, after which Get
// and List look them up by name. An empty name in Get picks the dark or
// light default from the terminal background.
package theme
