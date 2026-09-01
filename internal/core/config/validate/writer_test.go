package validate

import (
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

func TestWriterConfig(t *testing.T) {
	def := config.DefaultWriterConfig()

	tests := []struct {
		name          string
		mutateCfg     func(*config.WriterConfig)
		wantErrKeys   []string
		wantWarnCount int
		checkFixed    func(t *testing.T, cfg *config.WriterConfig)
	}{
		{
			name:          "valid config",
			mutateCfg:     func(c *config.WriterConfig) {},
			wantErrKeys:   nil,
			wantWarnCount: 0,
			checkFixed:    func(t *testing.T, c *config.WriterConfig) {},
		},
		{
			name: "MergeFlushInterval too low",
			mutateCfg: func(c *config.WriterConfig) {
				c.MergeFlushInterval = config.NewDurationMS(
					MinWriterMergeFlushInterval - 2*time.Millisecond,
				)
			},
			wantErrKeys:   []string{"MergeFlushInterval"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.WriterConfig) {
				if c.MergeFlushInterval.Duration() != def.MergeFlushInterval.Duration() {
					t.Errorf(
						"MergeFlushInterval = %v, want %v",
						c.MergeFlushInterval.Duration(),
						def.MergeFlushInterval.Duration(),
					)
				}
			},
		},
		{
			name: "MergeFlushInterval too high",
			mutateCfg: func(c *config.WriterConfig) {
				c.MergeFlushInterval = config.NewDurationMS(
					MaxWriterMergeFlushInterval + time.Second,
				)
			},
			wantErrKeys:   []string{"MergeFlushInterval"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.WriterConfig) {
				if c.MergeFlushInterval.Duration() != def.MergeFlushInterval.Duration() {
					t.Errorf(
						"MergeFlushInterval = %v, want %v",
						c.MergeFlushInterval.Duration(),
						def.MergeFlushInterval.Duration(),
					)
				}
			},
		},
		{
			name: "ChanSize too high",
			mutateCfg: func(c *config.WriterConfig) {
				c.ChanSize = MaxWriterChanSize + 1
			},
			wantErrKeys:   []string{"ChanSize"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.WriterConfig) {
				if c.ChanSize != def.ChanSize {
					t.Errorf(
						"ChanSize = %d, want %d",
						c.ChanSize,
						def.ChanSize,
					)
				}
			},
		},
		{
			name: "ChanSize too low",
			mutateCfg: func(c *config.WriterConfig) {
				c.ChanSize = MinWriterChanSize - 1
			},
			wantErrKeys:   []string{"ChanSize"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.WriterConfig) {
				if c.ChanSize != def.ChanSize {
					t.Errorf(
						"ChanSize = %d, want %d",
						c.ChanSize,
						def.ChanSize,
					)
				}
			},
		},
		{
			name: "BatchSize too high",
			mutateCfg: func(c *config.WriterConfig) {
				c.BatchSize = MaxWriterBatchSize + 1
			},
			wantErrKeys:   []string{"BatchSize"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.WriterConfig) {
				if c.BatchSize != def.BatchSize {
					t.Errorf(
						"BatchSize = %d, want %d",
						c.BatchSize,
						def.BatchSize,
					)
				}
			},
		},
		{
			name: "BatchSize too low",
			mutateCfg: func(c *config.WriterConfig) {
				c.BatchSize = MinWriterBatchSize - 1
			},
			wantErrKeys:   []string{"BatchSize"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.WriterConfig) {
				if c.BatchSize != def.BatchSize {
					t.Errorf(
						"BatchSize = %d, want %d",
						c.BatchSize,
						def.BatchSize,
					)
				}
			},
		},
		{
			name: "ResultBaseDir invalid",
			mutateCfg: func(c *config.WriterConfig) {
				c.ResultBaseDir = "invalid\\dir?"
			},
			wantErrKeys:   []string{"ResultBaseDir"},
			wantWarnCount: 1,
			checkFixed: func(t *testing.T, c *config.WriterConfig) {
				if c.ResultBaseDir != def.ResultBaseDir {
					t.Errorf(
						"ResultBaseDir = %q, want %q",
						c.ResultBaseDir,
						def.ResultBaseDir,
					)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := def
			tt.mutateCfg(&cfg)

			errs := ValidateWriter(cfg)

			if len(errs) != len(tt.wantErrKeys) {
				t.Errorf(
					"ValidateWriter() returned %d errors, want %d. Errors: %v",
					len(errs),
					len(tt.wantErrKeys),
					errs,
				)
			}

			for _, key := range tt.wantErrKeys {
				if _, ok := errs[key]; !ok {
					t.Errorf(
						"ValidateWriter() missing expected error for key %q",
						key,
					)
				}
			}

			cfg = def
			tt.mutateCfg(&cfg)

			warns := NormalizeWriter(&cfg)

			if len(warns) != tt.wantWarnCount {
				t.Errorf(
					"NormalizeWriter() returned %d warnings, want %d. Warnings: %v",
					len(warns),
					tt.wantWarnCount,
					warns,
				)
			}

			tt.checkFixed(t, &cfg)
		})
	}
}
