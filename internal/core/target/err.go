package target

import "errors"

// ErrTargetFileNotFound is returned when a named target list file does
// not exist.
var ErrTargetFileNotFound = errors.New("target file not found")
