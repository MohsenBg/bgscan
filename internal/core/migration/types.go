package migration

import (
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// TargetVersion is the canonical target-list layout: value,port,path,enable.
const TargetVersion = 2

// legacyIPDir is the V1 location; V1 lists lived directly under "ips/".
const legacyIPDir = "ips"

type targetVersion uint8

const (
	targetUnknown targetVersion = iota
	targetV1
	targetV2
)

// TargetReport describes what MigrateTargets changed.
type TargetReport struct {
	// Upgraded is the count of target files rewritten to the current layout.
	Upgraded int
}

// MigrationService upgrades on-disk state to the current layout.
type MigrationService interface {
	// MigrateTargets upgrades target lists to the current layout and
	// reports how many files were rewritten.
	MigrateTargets() (TargetReport, error)
}

type migrationService struct {
	baseDir string
}

// Option sets one field on the service created by New.
type Option func(*migrationService)

// WithBaseDir sets the base directory that holds the on-disk state.
// WithBaseDir trims spaces from dir.
func WithBaseDir(dir string) Option {
	return func(m *migrationService) {
		m.baseDir = strings.TrimSpace(dir)
	}
}

// New creates a MigrationService and applies opts in order.
// New uses the program base path when no base directory is set, and
// returns an error when that path cannot be found.
func New(opts ...Option) (MigrationService, error) {
	service := &migrationService{}
	for _, opt := range opts {
		opt(service)
	}

	if service.baseDir == "" {
		dir, err := fileutil.BasePath()
		if err != nil {
			return nil, err
		}
		service.baseDir = dir
	}

	return service, nil
}
