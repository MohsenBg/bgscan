package uisettings

import (
	"errors"
	"fmt"
	"math"
)

var (
	errPositiveRows = errors.New("max rows must be positive")
	errMaxRows      = fmt.Errorf("max rows cannot exceed %d", uint64(math.MaxUint32))
)
