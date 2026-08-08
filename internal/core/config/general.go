package config

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GeneralConfig defines global scanner behavior and execution settings.
type GeneralConfig struct {
	StatusInterval   DurationMS `toml:"status_interval"`
	StopAfterFound   int        `toml:"stop_after_found"`
	MaxIPsToTest     int        `toml:"max_ips_to_test"`
	PipelineMode     string     `toml:"pipeline_mode"`
	MaxIPsPerStage   int        `toml:"max_ips_per_stage"`
	BatchSize        int        `toml:"batch_size"`
	Shuffled         bool       `toml:"shuffled"`
	MinProbeDuration DurationMS `toml:"min_probe_duration"`
	ProbePerSec      int        `toml:"probe_per_sec"`
	ProbeBurst       int        `toml:"probe_burst"`
}

// platformProfile holds safe defaults per platform.
type platformProfile struct {
	minProbeDuration time.Duration
	probePerSec      int
	probeBurst       int
	maxIPsPerStage   int
	batchSize        int
}

// profiles defines conservative-but-functional defaults per platform.
// All values are non-zero — rate limiting is always active.
var profiles = map[string]platformProfile{
	"android_low": {
		minProbeDuration: 100 * time.Millisecond,
		probePerSec:      120,
		probeBurst:       30,
		maxIPsPerStage:   10_000,
		batchSize:        500,
	},
	"android_mid": {
		minProbeDuration: 80 * time.Millisecond,
		probePerSec:      160,
		probeBurst:       50,
		maxIPsPerStage:   50_000,
		batchSize:        2_000,
	},
	"android_high": {
		minProbeDuration: 50 * time.Millisecond,
		probePerSec:      200,
		probeBurst:       100,
		maxIPsPerStage:   100_000,
		batchSize:        5_000,
	},
	"desktop": {
		minProbeDuration: 50 * time.Millisecond,
		probePerSec:      500,
		probeBurst:       100,
		maxIPsPerStage:   100_000,
		batchSize:        5_000,
	},
}

// selectProfile picks a profile based on the current platform.
// On Android it reads /proc/meminfo to pick low/mid/high.
// Always returns a valid non-zero profile
func selectProfile() platformProfile {
	if runtime.GOOS == "android" {
		mb := availableMemoryMB()
		switch {
		case mb < 1500:
			return profiles["android_low"]
		case mb < 3500:
			return profiles["android_mid"]
		default:
			return profiles["android_high"]
		}
	}
	return profiles["desktop"]
}

// DefaultGeneralConfig returns the default configuration for general scanner behavior.
// Values are selected based on the current platform automatically.
func DefaultGeneralConfig() GeneralConfig {
	p := selectProfile()
	return GeneralConfig{
		StatusInterval:   NewDurationMS(1 * time.Second),
		StopAfterFound:   0,
		MaxIPsToTest:     0,
		MaxIPsPerStage:   p.maxIPsPerStage,
		BatchSize:        p.batchSize,
		Shuffled:         true,
		PipelineMode:     "streaming",
		MinProbeDuration: NewDurationMS(p.minProbeDuration),
		ProbePerSec:      p.probePerSec,
		ProbeBurst:       p.probeBurst,
	}
}

// availableMemoryMB reads total RAM from /proc/meminfo.
// Returns 4096 as a safe fallback if the file cannot be read.
func availableMemoryMB() int {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 4096 // fallback → android_high profile
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				break
			}
			kb, err := strconv.Atoi(fields[1])
			if err != nil {
				break
			}
			return kb / 1024
		}
	}
	return 4096
}
