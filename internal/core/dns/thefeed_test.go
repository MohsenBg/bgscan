package dns

import (
	"context"
	"net/netip"
	"testing"
	"time"
)

func TestTheFeedConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TheFeedConfig
		wantErr string
	}{
		{
			name: "valid config",
			config: TheFeedConfig{
				Domain:       "test.example.com",
				Passphrase:   "secret",
				ResolverPort: 5300,
				QueryMode:    "single",
			},
		},
		{
			name:    "empty domain",
			config:  TheFeedConfig{Passphrase: "secret"},
			wantErr: "domain",
		},
		{
			name:    "empty passphrase",
			config:  TheFeedConfig{Domain: "test.example.com"},
			wantErr: "passphrase",
		},
		{
			name:    "zero port",
			config:  TheFeedConfig{Domain: "test.example.com", Passphrase: "secret"},
			wantErr: "resolver_port",
		},
		{
			name:    "invalid query mode",
			config:  TheFeedConfig{Domain: "test.example.com", Passphrase: "secret", QueryMode: "invalid"},
			wantErr: "query_mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Validate()
			if len(errs) == 0 && tt.wantErr == "" {
				return
			}
			if len(errs) == 0 && tt.wantErr != "" {
				t.Fatalf("expected error for %q, got none", tt.wantErr)
			}
			if tt.wantErr != "" {
				found := false
				for k := range errs {
					if k == tt.wantErr {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected error key %q, got %v", tt.wantErr, errs)
				}
			}
		})
	}
}

func TestTheFeedService_Defaults(t *testing.T) {
	svc := NewTheFeedService()
	if svc == nil {
		t.Fatal("NewTheFeedService() returned nil")
	}

	_, err := svc.LoadConfig("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent config")
	}

	// ConfigDir returns the correct directory path
	files, err := svc.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles() error: %v", err)
	}
	if files == nil {
		t.Fatal("GetAllConfigFiles() returned nil")
	}
}

func TestTheFeedProbeResolverNeedsServer(t *testing.T) {
	svc := NewTheFeedService()

	// Without a running thefeed server, this should fail to connect
	// but not panic or crash.
	err := svc.RunTunnel(context.Background(), TheFeedConfig{
		Domain:       "nonexistent.thefeed.local",
		Passphrase:   "test",
		ResolverPort: 5300,
	}, netip.MustParseAddr("127.0.0.1"), 2*time.Second)

	// Should return an error (no server running)
	if err == nil {
		t.Fatal("expected error when probing without server")
	}
}
