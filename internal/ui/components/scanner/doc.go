// Package scanner renders the live scan screen. It shows one tab per scan
// stage, a progress bar with rate and ETA, Scan and Errors summary panels,
// and a table of results as they arrive. Engine hooks feed results and
// progress into the model from worker goroutines. The model also owns
// pausing, closing, and error dialogs for the run.
package scanner
