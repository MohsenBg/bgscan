package slipstream

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"unsafe"

	"github.com/MohsenBg/bgscan/internal/core/ffi"
	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/ebitengine/purego"
)

type DNSResolution int32

const (
	ModeRecursive     DNSResolution = 1
	ModeAuthoritative DNSResolution = 2
)

type Resolver struct {
	Host string
	Port uint16
	Mode DNSResolution
}

type Config struct {
	ListenHost        string
	ListenPort        uint16
	Resolvers         []Resolver
	Domain            string
	CongestionControl string
	GSO               bool
	CertPath          string
	KeepAliveInterval uint16
	DebugPoll         bool
	DebugStreams      bool
}

// C-layout mirrors. Field order/types must match the header exactly.
type cResolver struct {
	Host *byte
	Port uint16
	Mode int32
}

type cConfig struct {
	TCPListenHost     *byte
	TCPListenPort     uint16
	Resolvers         *cResolver
	ResolverCount     uintptr
	Domain            *byte
	CongestionControl *byte
	GSO               bool
	CertPath          *byte
	KeepAliveInterval uint16
	DebugPoll         bool
	DebugStreams      bool
}

// Lib is a loaded libslipstream.
type Lib struct {
	version func() uintptr
	start   func(cfg unsafe.Pointer) uintptr
	stop    func(client uintptr) int32
	running func(client uintptr) bool
}

// Client represents a running Slipstream tunnel.
type Client interface {
	Running() bool
	Stop() error
}

// client is the concrete Slipstream client implementation.
type client struct {
	lib *Lib

	mu     sync.Mutex
	handle uintptr

	// Keeps all C-visible Go memory alive until Stop.
	keep []any
}

// Load opens the shared library and binds its functions.
func Load() (*Lib, error) {
	h, err := loadLibSlipstream()
	if err != nil {
		return nil, err
	}

	l := &Lib{}

	purego.RegisterLibFunc(
		&l.version,
		h,
		"slipstream_version",
	)

	purego.RegisterLibFunc(
		&l.start,
		h,
		"slipstream_client_start",
	)

	purego.RegisterLibFunc(
		&l.stop,
		h,
		"slipstream_client_stop",
	)

	purego.RegisterLibFunc(
		&l.running,
		h,
		"slipstream_client_is_running",
	)

	return l, nil
}

func (l *Lib) Version() string {
	return ffi.GoString(l.version())
}

// Start creates and starts a client.
func (l *Lib) Start(cfg Config) (Client, error) {
	if len(cfg.Resolvers) == 0 {
		return nil, errors.New("at least one resolver is required")
	}

	var keep []any

	str := func(s string) *byte {
		b := ffi.CString(s)
		keep = append(keep, b)
		return &b[0]
	}

	cRes := make([]cResolver, len(cfg.Resolvers))

	for i, r := range cfg.Resolvers {
		cRes[i] = cResolver{
			Host: str(r.Host),
			Port: r.Port,
			Mode: int32(r.Mode),
		}
	}

	keep = append(keep, cRes)

	cc := &cConfig{
		TCPListenHost:     str(cfg.ListenHost),
		TCPListenPort:     cfg.ListenPort,
		Resolvers:         &cRes[0],
		ResolverCount:     uintptr(len(cRes)),
		Domain:            str(cfg.Domain),
		CongestionControl: ffi.OptionalCString(cfg.CongestionControl, &keep),
		GSO:               cfg.GSO,
		CertPath:          ffi.OptionalCString(cfg.CertPath, &keep),
		KeepAliveInterval: cfg.KeepAliveInterval,
		DebugPoll:         cfg.DebugPoll,
		DebugStreams:      cfg.DebugStreams,
	}

	keep = append(keep, cc)

	handle := l.start(unsafe.Pointer(cc))
	if handle == 0 {
		return nil, errors.New("slipstream_client_start failed")
	}

	return &client{
		lib:    l,
		handle: handle,
		keep:   keep,
	}, nil
}

// Running reports whether the client is still running.
// Returns false after Stop.
func (c *client) Running() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handle == 0 {
		return false
	}

	return c.lib.running(c.handle)
}

// Stop stops and frees the client.
// Stop is idempotent.
func (c *client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handle == 0 {
		return nil
	}

	rc := c.lib.stop(c.handle)

	c.handle = 0
	c.keep = nil

	if rc != 0 {
		return fmt.Errorf(
			"slipstream_client_stop returned %d",
			rc,
		)
	}

	return nil
}

func (c *client) Kill() error {
	return c.Stop()
}

// getSlipstreamPaths returns the directories searched for libslipstream.
func getSlipstreamPaths() []string {
	base, err := fileutil.BasePath()
	if err != nil {
		return nil
	}

	return []string{
		filepath.Join(base, "assets", "slipstream-client"),
		filepath.Join(base, "assets", "slipstream"),
		filepath.Join(base, "assets"),
		base,
	}
}

// libFilePatterns returns per-platform filename patterns; versioned and
// arch-suffixed builds match via wildcards.
func libFilePatterns() []string {
	switch runtime.GOOS {
	case "linux", "android":
		return []string{
			"libslipstream-client.so",
			"libslipstream-client-*.so",
		}

	case "darwin":
		return []string{
			"libslipstream-client.dylib",
			"libslipstream-client-*.dylib",
		}

	case "windows":
		return []string{
			"slipstream-client.dll",
			"libslipstream-client.dll",
			"libslipstream-client-*.dll",
		}

	default:
		return nil
	}
}

// FindLibSlipstream locates libslipstream in the known locations.
func FindLibSlipstream() (string, error) {
	patterns := libFilePatterns()

	if len(patterns) == 0 {
		return "", fmt.Errorf(
			"unsupported platform: %s",
			runtime.GOOS,
		)
	}

	for _, dir := range getSlipstreamPaths() {
		var candidates []string

		for _, pattern := range patterns {
			matches, err := filepath.Glob(
				filepath.Join(dir, pattern),
			)
			if err != nil {
				continue
			}

			candidates = append(candidates, matches...)
		}

		if len(candidates) == 0 {
			continue
		}

		sort.Strings(candidates)

		for _, path := range candidates {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}

			return path, nil
		}
	}

	return "", fmt.Errorf(
		"slipstream-client library not found",
	)
}

// VerifyLibSlipstream loads the library and reports its version.
func VerifyLibSlipstream() (string, error) {
	lib, err := Load()
	if err != nil {
		return "", err
	}

	version := lib.Version()

	if version == "" {
		return "", fmt.Errorf(
			"slipstream library returned empty version",
		)
	}

	return version, nil
}

func loadLibSlipstream() (uintptr, error) {
	path, err := FindLibSlipstream()
	if err != nil {
		return 0, fmt.Errorf(
			"find lib-slipstream-client: %w",
			err,
		)
	}

	return ffi.OpenLibrary(path)
}
