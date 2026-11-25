// Package sparkline provides an inline chart component for the terminal
// UI. It draws samples as braille cells with two samples per cell and
// keeps only the newest samples that fit the chart width. Callers add
// data with Push or SetSamples and read the rendered chart with View.
package sparkline
