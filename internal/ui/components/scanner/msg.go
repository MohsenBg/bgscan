package scanner

import (
	"github.com/MohsenBg/bgscan/internal/core/scanner/engine"
)

type scanErrorMsg struct {
	err error
}

type scanCloseResultMsg struct {
	err error
}

type stageProgressMsg struct {
	id int
	p  engine.Progress
}

type stageErrorMsg struct {
	id  int
	err error
}

type stageEndMsg struct{ id int }

type pulseMsg struct{}
