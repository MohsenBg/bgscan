// Package slipstream loads the native libslipstream library and starts
// Slipstream tunnel clients through it.
//
// The main types are Lib, which holds the bound library functions, and
// Client, which is a running tunnel. Callers load the library with Load,
// start a client with Lib.Start, and must call Client.Stop when done so
// the native client is freed.
package slipstream
