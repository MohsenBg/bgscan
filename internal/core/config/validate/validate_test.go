package validate

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"bgscan/internal/core/config"
)

func TestWarningString(t *testing.T) {
	w := Warning{Field: "testField", OldVal: 10, NewVal: 20, Reason: "out of range"}
	if got := w.String(); got != "testField: 10 → 20 (out of range)" {
		t.Errorf("Warning.String() = %q, want %q", got, "testField: 10 → 20 (out of range)")
	}
}

func TestCheckAndFixInt(t *testing.T) {
	tests := []struct {
		name               string
		val, min, max, def int
		wantErr            error
		wantFixed          int
	}{
		{"valid", 5, 1, 10, 0, nil, 5},
		{"below min", 0, 1, 10, 42, ErrOutOfRange, 42},
		{"above max", 11, 1, 10, 42, ErrOutOfRange, 42},
		{"exact min", 1, 1, 10, 0, nil, 1},
		{"exact max", 10, 1, 10, 0, nil, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkInt()", checkInt("testField", tt.val, tt.min, tt.max), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixInt("testField", &v, tt.min, tt.max, tt.def, &warns)

			if v != tt.wantFixed {
				t.Errorf("fixInt() val = %v, want %v", v, tt.wantFixed)
			}
			checkWarned(t, "fixInt()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixUint16(t *testing.T) {
	tests := []struct {
		name               string
		val, min, max, def uint16
		wantErr            error
		wantFixed          uint16
	}{
		{"valid", 5, 1, 10, 0, nil, 5},
		{"below min", 0, 1, 10, 42, ErrOutOfRange, 42},
		{"above max", 11, 1, 10, 42, ErrOutOfRange, 42},
		{"exact min", 1, 1, 10, 0, nil, 1},
		{"exact max", 10, 1, 10, 0, nil, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkUint16()", checkUint16("testField", tt.val, tt.min, tt.max), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixUint16("testField", &v, tt.min, tt.max, tt.def, &warns)

			if v != tt.wantFixed {
				t.Errorf("fixUint16() val = %v, want %v", v, tt.wantFixed)
			}
			checkWarned(t, "fixUint16()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixString(t *testing.T) {
	tests := []struct {
		name      string
		val, def  string
		wantErr   error
		wantFixed string
	}{
		{"valid", "hello", "", nil, "hello"},
		{"empty", "", "default", ErrEmpty, "default"},
		{"whitespace only", "   ", "default", ErrEmpty, "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkString()", checkString("testField", tt.val), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixString("testField", &v, tt.def, &warns)

			if v != tt.wantFixed {
				t.Errorf("fixString() val = %q, want %q", v, tt.wantFixed)
			}
			checkWarned(t, "fixString()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixStringSlice(t *testing.T) {
	tests := []struct {
		name      string
		val, def  []string
		wantErr   error
		wantFixed []string
	}{
		{"valid", []string{"a", "b"}, nil, nil, []string{"a", "b"}},
		{"empty slice", []string{}, []string{"default"}, ErrEmpty, []string{"default"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkStringSlice()", checkStringSlice("testField", tt.val), tt.wantErr)

			v := make([]string, len(tt.val))
			copy(v, tt.val)
			var warns []Warning
			fixStringSlice("testField", &v, tt.def, &warns)

			if !reflect.DeepEqual(v, tt.wantFixed) {
				t.Errorf("fixStringSlice() val = %v, want %v", v, tt.wantFixed)
			}
			checkWarned(t, "fixStringSlice()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixEnum(t *testing.T) {
	allowed := []string{"a", "b", "c"}
	tests := []struct {
		name      string
		val, def  string
		wantErr   error
		wantFixed string
	}{
		{"valid", "b", "", nil, "b"},
		{"valid upper", "A", "", nil, "A"},
		{"invalid", "d", "a", ErrInvalidEnum, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkEnum()", checkEnum("testField", tt.val, allowed), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixEnum("testField", &v, allowed, tt.def, &warns)

			if v != tt.wantFixed {
				t.Errorf("fixEnum() val = %q, want %q", v, tt.wantFixed)
			}
			checkWarned(t, "fixEnum()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixEnumOrder(t *testing.T) {
	allowed := []string{"low", "medium", "high"}
	tests := []struct {
		name           string
		minVal, maxVal string
		minDef, maxDef string
		wantErr        error
		wantMinFix     string
		wantMaxFix     string
	}{
		{"valid order", "low", "high", "low", "high", nil, "low", "high"},
		{"same level", "medium", "medium", "low", "high", nil, "medium", "medium"},
		{"invalid order", "high", "low", "low", "high", ErrEnumOrder, "low", "high"},
		{"invalid values ignored", "invalid1", "invalid2", "low", "high", nil, "invalid1", "invalid2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkEnumOrder()", checkEnumOrder("minField", "maxField", tt.minVal, tt.maxVal, allowed), tt.wantErr)

			minV, maxV := tt.minVal, tt.maxVal
			var warns []Warning
			fixEnumOrder("minField", "maxField", &minV, &maxV, tt.minDef, tt.maxDef, allowed, &warns)

			if minV != tt.wantMinFix || maxV != tt.wantMaxFix {
				t.Errorf("fixEnumOrder() vals = %q, %q, want %q, %q", minV, maxV, tt.wantMinFix, tt.wantMaxFix)
			}
			checkWarned(t, "fixEnumOrder()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixPrefix(t *testing.T) {
	tests := []struct {
		name     string
		val, def string
		wantErr  error
	}{
		{"valid", "my-prefix", "", nil},
		{"empty", "", "default", ErrEmpty},
		{"illegal char", "my:prefix", "default", ErrIllegalChars},
		{"leading dot", ".hidden", "default", ErrLeadingTrailing},
		{"trailing space", "trailing ", "default", ErrLeadingTrailing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkPrefix()", checkPrefix("testField", tt.val), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixPrefix("testField", &v, tt.def, &warns)

			if tt.wantErr != nil && v != tt.def {
				t.Errorf("fixPrefix() val = %q, want %q", v, tt.def)
			}
		})
	}
}

func TestCheckAndFixDirectoryName(t *testing.T) {
	tests := []struct {
		name     string
		val, def string
		wantErr  error
	}{
		{"valid", "mydir", "", nil},
		{"empty/whitespace", "   ", "default", ErrEmpty},
		{"dot", ".", "default", ErrIllegalChars},
		{"dotdot", "..", "default", ErrIllegalChars},
		{"path traversal", "a/../b", "default", ErrIllegalChars},
		{"illegal char", "dir?", "default", ErrIllegalChars},
		{"leading dot", ".hidden", "default", ErrLeadingTrailing},
		{"trailing space", "dir ", "default", ErrLeadingTrailing},
		{"contains ..", "dir..name", "default", ErrDotDot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkDirectoryName()", checkDirectoryName("testField", tt.val), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixDirectoryName("testField", &v, tt.def, &warns)

			if tt.wantErr != nil && v != tt.def {
				t.Errorf("fixDirectoryName() val = %q, want %q", v, tt.def)
			}
		})
	}
}

func TestCheckAndFixHost(t *testing.T) {
	tests := []struct {
		name     string
		val, def string
		wantErr  error
	}{
		{"valid domain", "google.com", "", nil},
		{"with scheme", "http://google.com", "default.com", ErrSchemeNotAllowed},
		{"with uppercase scheme", "HTTPS://google.com", "default.com", ErrSchemeNotAllowed},
		{"with port and path", "example.com:8080/api/v1", "", nil},
		{"invalid domain", "invalid_domain!", "default.com", ErrInvalidDomain},
		{"empty", "", "default.com", ErrEmpty},
		{"url parse error", "google.com\x00", "default.com", ErrMalformedURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkHost()", checkHost("testField", tt.val), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixHost("testField", &v, tt.def, &warns)

			if tt.wantErr != nil && v != tt.def {
				t.Errorf("fixHost() val = %q, want %q", v, tt.def)
			}
		})
	}
}

func TestCheckAndFixSNI(t *testing.T) {
	tests := []struct {
		name     string
		val, def string
		wantErr  error
	}{
		{"valid domain", "google.com", "", nil},
		{"with scheme", "https://google.com", "default.com", ErrSchemeNotAllowed},
		{"with uppercase scheme", "HTTP://google.com", "default.com", ErrSchemeNotAllowed},
		{"with path", "google.com/path", "default.com", ErrPathOrPort},
		{"with port", "google.com:443", "default.com", ErrPathOrPort},
		{"empty is valid", "", "", nil},
		{"numeric TLD", "google.123", "default.com", ErrInvalidDomain},
		{"invalid starting char", "-invalid.com", "default.com", ErrInvalidDomain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkSNI()", checkSNI("testField", tt.val), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixSNI("testField", &v, tt.def, &warns)

			if tt.wantErr != nil && v != tt.def {
				t.Errorf("fixSNI() val = %q, want %q", v, tt.def)
			}
		})
	}
}

func TestCheckAndFixDomain(t *testing.T) {
	tests := []struct {
		name     string
		val, def string
		wantErr  error
	}{
		{"valid domain", "google.com", "", nil},
		{"with scheme", "https://google.com", "default.com", ErrSchemeNotAllowed},
		{"with uppercase scheme", "Http://google.com", "default.com", ErrSchemeNotAllowed},
		{"with path", "google.com/path", "default.com", ErrPathOrPort},
		{"with port", "google.com:443", "default.com", ErrPathOrPort},
		{"empty is not valid", "", "", ErrEmpty},
		{"numeric TLD", "google.123", "default.com", ErrInvalidDomain},
		{"invalid starting char", "-invalid.com", "default.com", ErrInvalidDomain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkDomain()", checkDomain("testField", tt.val), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixDomain("testField", &v, tt.def, &warns)

			if tt.wantErr != nil && v != tt.def {
				t.Errorf("fixDomain() val = %q, want %q", v, tt.def)
			}
		})
	}
}

func TestCheckAndFixDurationMS(t *testing.T) {
	minDur := 1 * time.Second
	maxDur := 5 * time.Second
	defDur := 2 * time.Second

	tests := []struct {
		name      string
		val       time.Duration
		wantErr   error
		wantFixed time.Duration
	}{
		{"valid", 3 * time.Second, nil, 3 * time.Second},
		{"exact min", 1 * time.Second, nil, 1 * time.Second},
		{"exact max", 5 * time.Second, nil, 5 * time.Second},
		{"just below min (within epsilon)", 999 * time.Millisecond, nil, 999 * time.Millisecond},
		{"below min (outside epsilon)", 998 * time.Millisecond, ErrOutOfRange, defDur},
		{"just above max (within epsilon)", 5001 * time.Millisecond, nil, 5001 * time.Millisecond},
		{"above max (outside epsilon)", 5002 * time.Millisecond, ErrOutOfRange, defDur},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := config.NewDurationMS(tt.val)

			checkErr(t, "checkDuration()", checkDuration("testField", v.Duration(), minDur, maxDur), tt.wantErr)

			var warns []Warning
			fixDurationMS("testField", &v, minDur, maxDur, config.NewDurationMS(defDur), &warns)

			if v.Duration() != tt.wantFixed {
				t.Errorf("fixDurationMS() val = %v, want %v", v.Duration(), tt.wantFixed)
			}
			checkWarned(t, "fixDurationMS()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixStatusCodes(t *testing.T) {
	tests := []struct {
		name      string
		val, def  []int
		wantErr   error
		wantFixed []int
	}{
		{"valid", []int{200, 404}, nil, nil, []int{200, 404}},
		{"invalid code", []int{200, 999}, []int{200}, ErrInvalidStatusCode, []int{200}},
		{"empty slice", []int{}, []int{200}, nil, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkStatusCodes()", checkStatusCodes("testField", tt.val), tt.wantErr)

			v := make([]int, len(tt.val))
			copy(v, tt.val)
			var warns []Warning
			fixHTTPStatusCodes("testField", &v, tt.def, &warns)

			if tt.wantErr != nil && !reflect.DeepEqual(v, tt.wantFixed) {
				t.Errorf("fixHTTPStatusCodes() val = %v, want %v", v, tt.wantFixed)
			}
			checkWarned(t, "fixHTTPStatusCodes()", warns, tt.wantErr)
		})
	}
}

func TestCheckAndFixPubKey(t *testing.T) {
	validKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	tests := []struct {
		name     string
		val, def string
		wantErr  error
	}{
		{"valid", validKey, "", nil},
		{"too short", "abc", "default", ErrInvalidPubKey},
		{"invalid chars", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b!", "default", ErrInvalidPubKey},
		{"empty", "", "default", ErrEmpty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr(t, "checkPubKey()", checkPubKey("testField", tt.val), tt.wantErr)

			v := tt.val
			var warns []Warning
			fixPubKey("testField", &v, tt.def, &warns)

			if tt.wantErr != nil && v != tt.def {
				t.Errorf("fixPubKey() val = %q, want %q", v, tt.def)
			}
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"google.com", true},
		{"sub.example.com", true},
		{"localhost", true},
		{"192.168.1.1", true},
		{"", false},
		{"-google.com", false},
		{"google.c", true},
		{"google.123", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isValidDomain(tt.host); got != tt.want {
				t.Errorf("isValidDomain(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestIsValidHTTPStatusCode(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{99, false},
		{100, true},
		{200, true},
		{599, true},
		{600, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := isValidHTTPStatusCode(tt.code); got != tt.want {
				t.Errorf("isValidHTTPStatusCode(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

// checkErr asserts that err wraps wantErr. If wantErr is nil, err must be nil.
func checkErr(t *testing.T, fn string, err error, wantErr error) {
	t.Helper()
	if wantErr == nil {
		if err != nil {
			t.Errorf("%s unexpected error: %v", fn, err)
		}
		return
	}
	if err == nil {
		t.Errorf("%s expected error wrapping %v, got nil", fn, wantErr)
		return
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("%s error = %v, want errors.Is(%v)", fn, err, wantErr)
	}
}

// checkWarned asserts that warns is non-empty iff wantErr is non-nil.
func checkWarned(t *testing.T, fn string, warns []Warning, wantErr error) {
	t.Helper()
	if wantErr != nil && len(warns) == 0 {
		t.Errorf("%s expected a warning, got none", fn)
	}
	if wantErr == nil && len(warns) > 0 {
		t.Errorf("%s unexpected warning: %v", fn, warns)
	}
}
