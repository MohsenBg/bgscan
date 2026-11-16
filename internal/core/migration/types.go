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
	Upgraded int
}

// MigrationService upgrades on-disk state to the current layout.
type MigrationService interface {
	MigrateTargets() (TargetReport, error)
}

type migrationService struct {
	baseDir string
}

type Option func(*migrationService)

func WithBaseDir(dir string) Option {
	return func(m *migrationService) {
		m.baseDir = strings.TrimSpace(dir)
	}
}

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
