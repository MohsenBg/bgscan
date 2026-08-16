package config

import (
	"errors"
	"os"
	"testing"
)

type fakeEnv struct {
	env   map[string]string
	files map[string][]byte
}

func (f *fakeEnv) Getenv(key string) string {
	return f.env[key]
}

func (f *fakeEnv) Stat(name string) (os.FileInfo, error) {
	if _, ok := f.files[name]; ok {
		return nil, nil
	}

	return nil, os.ErrNotExist
}

func (f *fakeEnv) ReadFile(name string) ([]byte, error) {
	data, ok := f.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	return data, nil
}

func newFakeEnv() *fakeEnv {
	return &fakeEnv{
		env:   make(map[string]string),
		files: make(map[string][]byte),
	}
}

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name string
		goos string
		env  *fakeEnv
		want Platform
	}{
		{
			name: "android",
			goos: "android",
			env:  newFakeEnv(),
			want: Android,
		},
		{
			name: "linux termux by environment",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.env["TERMUX_VERSION"] = "0.119"
				return env
			}(),
			want: Android,
		},
		{
			name: "linux termux by path",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.files["/data/data/com.termux/files/usr"] = nil
				return env
			}(),
			want: Android,
		},
		{
			name: "linux desktop with x11",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.env["DISPLAY"] = ":0"
				return env
			}(),
			want: Desktop,
		},
		{
			name: "linux desktop with wayland",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.env["WAYLAND_DISPLAY"] = "wayland-0"
				return env
			}(),
			want: Desktop,
		},
		{
			name: "linux server docker",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.files["/.dockerenv"] = nil
				return env
			}(),
			want: Server,
		},
		{
			name: "linux server systemd",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.files["/proc/1/comm"] = []byte("systemd\n")
				return env
			}(),
			want: Server,
		},
		{
			name: "linux without server signals",
			goos: "linux",
			env:  newFakeEnv(),
			want: Desktop,
		},
		{
			name: "windows",
			goos: "windows",
			env:  newFakeEnv(),
			want: Desktop,
		},
		{
			name: "darwin",
			goos: "darwin",
			env:  newFakeEnv(),
			want: Desktop,
		},
		{
			name: "unknown",
			goos: "freebsd",
			env:  newFakeEnv(),
			want: Desktop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectPlatform(
				WithGOOS(tt.goos),
				WithEnvReader(tt.env),
			)

			if got != tt.want {
				t.Fatalf("DetectPlatform() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectPlatformNilOptions(t *testing.T) {
	got := DetectPlatform(
		WithGOOS("android"),
		WithEnvReader(nil),
	)

	if got != Android {
		t.Fatalf("DetectPlatform() = %q, want %q", got, Android)
	}
}

func TestIsTermux(t *testing.T) {
	tests := []struct {
		name string
		env  *fakeEnv
		want bool
	}{
		{
			name: "environment variable",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.env["TERMUX_VERSION"] = "0.119"
				return env
			}(),
			want: true,
		},
		{
			name: "termux path",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.files["/data/data/com.termux/files/usr"] = nil
				return env
			}(),
			want: true,
		},
		{
			name: "not termux",
			env:  newFakeEnv(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTermux(tt.env); got != tt.want {
				t.Fatalf("isTermux() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsServer(t *testing.T) {
	tests := []struct {
		name string
		goos string
		env  *fakeEnv
		want bool
	}{
		{
			name: "non linux",
			goos: "windows",
			env:  newFakeEnv(),
			want: false,
		},
		{
			name: "x11 desktop",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.env["DISPLAY"] = ":0"
				return env
			}(),
			want: false,
		},
		{
			name: "wayland desktop",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.env["WAYLAND_DISPLAY"] = "wayland-0"
				return env
			}(),
			want: false,
		},
		{
			name: "docker",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.files["/.dockerenv"] = nil
				return env
			}(),
			want: true,
		},
		{
			name: "systemd",
			goos: "linux",
			env: func() *fakeEnv {
				env := newFakeEnv()
				env.files["/proc/1/comm"] = []byte("systemd\n")
				return env
			}(),
			want: true,
		},
		{
			name: "unknown linux",
			goos: "linux",
			env:  newFakeEnv(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isServer(tt.goos, tt.env); got != tt.want {
				t.Fatalf("isServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectTier(t *testing.T) {
	tests := []struct {
		name string
		res  SystemResources
		want Tier
	}{
		{
			name: "low cpu",
			res: SystemResources{
				CPUCores: 2,
				FDLimit:  4096,
			},
			want: Low,
		},
		{
			name: "low fd limit",
			res: SystemResources{
				CPUCores: 8,
				FDLimit:  511,
			},
			want: Low,
		},
		{
			name: "mid cpu",
			res: SystemResources{
				CPUCores: 8,
				FDLimit:  4096,
			},
			want: Mid,
		},
		{
			name: "mid fd limit",
			res: SystemResources{
				CPUCores: 16,
				FDLimit:  4095,
			},
			want: Mid,
		},
		{
			name: "high",
			res: SystemResources{
				CPUCores: 16,
				FDLimit:  4096,
			},
			want: High,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectTier(tt.res); got != tt.want {
				t.Fatalf("SelectTier() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsServerSystemdRequiresExactValue(t *testing.T) {
	env := newFakeEnv()
	env.files["/proc/1/comm"] = []byte("systemd-user\n")

	if isServer("linux", env) {
		t.Fatal("isServer() = true, want false")
	}
}

func TestFakeEnvMissingFile(t *testing.T) {
	env := newFakeEnv()

	_, err := env.ReadFile("/missing")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ReadFile() error = %v, want os.ErrNotExist", err)
	}
}
