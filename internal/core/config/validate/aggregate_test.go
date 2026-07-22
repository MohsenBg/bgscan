package validate

import (
	"testing"

	"bgscan/internal/core/config"
)

// ============================================================================
// AllWarnings Tests
// ============================================================================
func TestAllWarningsHasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		warnings AllWarnings
		want     bool
	}{
		{
			name:     "all empty",
			warnings: AllWarnings{},
			want:     false,
		},
		{
			name: "one section has warnings",
			warnings: AllWarnings{
				ICMP: []Warning{{Field: "Workers", OldVal: 0, NewVal: 1, Reason: "negative → default"}},
			},
			want: true,
		},
		{
			name: "multiple sections have warnings",
			warnings: AllWarnings{
				General: []Warning{{Field: "StatusInterval", OldVal: 0, NewVal: 5000, Reason: "too low → default"}},
				TCP:     []Warning{{Field: "Port", OldVal: 0, NewVal: 80, Reason: "invalid → default"}},
				DNS:     []Warning{{Field: "Resolver.Workers", OldVal: 0, NewVal: 100, Reason: "too low → default"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.warnings.HasWarnings(); got != tt.want {
				t.Errorf("AllWarnings.HasWarnings() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// AllErrors Tests
// ============================================================================
func TestAllErrorsHasErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors AllErrors
		want   bool
	}{
		{
			name:   "all nil/empty",
			errors: AllErrors{},
			want:   false,
		},
		{
			name: "one section has errors",
			errors: AllErrors{
				HTTP: map[string]error{"Port": errTest},
			},
			want: true,
		},
		{
			name: "multiple sections have errors",
			errors: AllErrors{
				Xray:   map[string]error{"Workers": errTest},
				Writer: map[string]error{"ChanSize": errTest},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.HasErrors(); got != tt.want {
				t.Errorf("AllErrors.HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Dummy error for testing map population
var errTest = &testError{s: "test error"}

type testError struct{ s string }

func (e *testError) Error() string { return e.s }

// ============================================================================
// ValidateAll Integration Test
// ============================================================================
func TestValidateAll(t *testing.T) {
	// Save original state
	origICMP := *config.GetICMP()
	origTCP := *config.GetTCP()

	// Ensure state is restored after test
	t.Cleanup(func() {
		*config.GetICMP() = origICMP
		*config.GetTCP() = origTCP
	})

	// 1. Test with valid global config (should have no errors)
	errs := ValidateAll()
	if errs.HasErrors() {
		t.Errorf("ValidateAll() on valid config returned errors: %v", errs)
	}

	// 2. Mutate global config to be invalid in specific sections
	config.GetICMP().Workers = 0 // Triggers ICMP error
	config.GetTCP().Port = 70000 // Triggers TCP error

	// 3. Test aggregation
	errs = ValidateAll()
	if !errs.HasErrors() {
		t.Fatal("ValidateAll() expected errors, got none")
	}

	if len(errs.ICMP) == 0 {
		t.Errorf("ValidateAll() missing expected ICMP errors, got: %v", errs.ICMP)
	}
	if _, ok := errs.ICMP["Workers"]; !ok {
		t.Errorf("ValidateAll() ICMP errors missing 'Workers' key, got: %v", errs.ICMP)
	}

	if len(errs.TCP) == 0 {
		t.Errorf("ValidateAll() missing expected TCP errors, got: %v", errs.TCP)
	}
	if _, ok := errs.TCP["Port"]; !ok {
		t.Errorf("ValidateAll() TCP errors missing 'Port' key, got: %v", errs.TCP)
	}

	// Verify other sections remain error-free (e.g., General)
	if len(errs.General) > 0 {
		t.Errorf("ValidateAll() unexpected General errors: %v", errs.General)
	}
}

// ============================================================================
// NormalizeAll Integration Test
// ============================================================================
func TestNormalizeAll(t *testing.T) {
	// Save original state
	origWriter := *config.GetWriter()
	origXray := *config.GetXray()

	// Ensure state is restored after test
	t.Cleanup(func() {
		*config.GetWriter() = origWriter
		*config.GetXray() = origXray
	})

	defWriter := config.DefaultWriterConfig()
	defXray := config.DefaultXrayConfig()

	// 1. Test with valid global config (should have no warnings)
	warns := NormalizeAll()
	if warns.HasWarnings() {
		t.Errorf("NormalizeAll() on valid config returned warnings: %v", warns)
	}

	// 2. Mutate global config to be invalid in specific sections
	config.GetWriter().ChanSize = 2_000_000 // Exceeds max 1,000,000
	config.GetXray().PreScanType = "invalid_mode"

	// 3. Test aggregation and auto-fix
	warns = NormalizeAll()
	if !warns.HasWarnings() {
		t.Fatal("NormalizeAll() expected warnings, got none")
	}

	if len(warns.Writer) == 0 {
		t.Errorf("NormalizeAll() missing expected Writer warnings, got: %v", warns.Writer)
	}
	if config.GetWriter().ChanSize != defWriter.ChanSize {
		t.Errorf("NormalizeAll() failed to fix Writer.ChanSize, got %d, want %d",
			config.GetWriter().ChanSize, defWriter.ChanSize)
	}

	if len(warns.Xray) == 0 {
		t.Errorf("NormalizeAll() missing expected Xray warnings, got: %v", warns.Xray)
	}
	if config.GetXray().PreScanType != defXray.PreScanType {
		t.Errorf("NormalizeAll() failed to fix Xray.PreScanType, got %q, want %q",
			config.GetXray().PreScanType, defXray.PreScanType)
	}

	// Verify other sections remain unmodified (e.g., General)
	// (We didn't mutate General, so it shouldn't generate warnings)
	if len(warns.General) > 0 {
		t.Errorf("NormalizeAll() unexpected General warnings: %v", warns.General)
	}
}
