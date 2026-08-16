package config

import (
	"os"
	"runtime"
	"strings"
)

// Platform represents the target platform type.
type Platform string

const (
	Android Platform = "android"
	Desktop Platform = "desktop"
	Server  Platform = "server"
)

// Tier represents the resource tier for configuration defaults.
type Tier string

const (
	Low  Tier = "low"
	Mid  Tier = "mid"
	High Tier = "high"
)

// SystemResources holds detected system capabilities.
type SystemResources struct {
	CPUCores int
	FDLimit  uint64
}

// EnvReader abstracts environment access for testing.
type EnvReader interface {
	Getenv(key string) string
	Stat(name string) (os.FileInfo, error)
	ReadFile(name string) ([]byte, error)
}

type osEnv struct{}

func (o osEnv) Getenv(key string) string              { return os.Getenv(key) }
func (o osEnv) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (o osEnv) ReadFile(name string) ([]byte, error)  { return os.ReadFile(name) }

type Option func(*detectOptions)

type detectOptions struct {
	env  EnvReader
	goos string
}

func WithEnvReader(env EnvReader) Option {
	return func(o *detectOptions) {
		if env != nil {
			o.env = env
		}
	}
}

func WithGOOS(goos string) Option {
	return func(o *detectOptions) {
		o.goos = goos
	}
}

// DetectPlatform determines the platform from GOOS and environment.
// Linux systems are further classified as Server, Android (Termux), or Desktop.
func DetectPlatform(opts ...Option) Platform {
	cfg := detectOptions{
		env:  osEnv{},
		goos: runtime.GOOS,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	switch cfg.goos {
	case "android":
		return Android

	case "linux":
		if isTermux(cfg.env) {
			return Android
		}
		if isServer(cfg.goos, cfg.env) {
			return Server
		}
		return Desktop

	default:
		return Desktop
	}
}

func isServer(goos string, env EnvReader) bool {
	if goos != "linux" {
		return false
	}

	if env.Getenv("DISPLAY") != "" || env.Getenv("WAYLAND_DISPLAY") != "" {
		return false
	}

	if _, err := env.Stat("/.dockerenv"); err == nil {
		return true
	}

	if pid1, err := env.ReadFile("/proc/1/comm"); err == nil {
		if strings.TrimSpace(string(pid1)) == "systemd" {
			return true
		}
	}

	return false
}

func isTermux(env EnvReader) bool {
	if env.Getenv("TERMUX_VERSION") != "" {
		return true
	}

	_, err := env.Stat("/data/data/com.termux/files/usr")
	return err == nil
}

// CheckResources detects CPU cores and file descriptor limit.
func CheckResources() SystemResources {
	res := SystemResources{
		CPUCores: runtime.NumCPU(),
		FDLimit:  getFDLimit(),
	}
	return res
}

// SelectTier chooses a resource tier based on system capabilities.
func SelectTier(res SystemResources) Tier {
	switch {
	case res.FDLimit < 512 || res.CPUCores <= 2:
		return Low
	case res.FDLimit < 4096 || res.CPUCores <= 8:
		return Mid
	default:
		return High
	}
}
