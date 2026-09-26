package uisettings

import (
	"errors"
	"fmt"
	"math"
)

var (
	errPositiveRows = errors.New("max rows must be a positive number")
	errMaxRows      = fmt.Errorf("max rows cannot exceed %d", math.MaxUint32)
)
