package result

import (
	"math"
	"testing"
	"time"
)

// ---------- KeyType ----------

func TestKeyType_String(t *testing.T) {
	tests := []struct {
		name string
		kt   KeyType
		want string
	}{
		{"ip", KeyIP, "ip"},
		{"domain", KeyDomain, "domain"},
		{"undefined high", KeyType(255), "unknown(255)"},
		{"zero", KeyType(0), "ip"}, // KeyIP == 0
		{"max valid", KeyDomain, "domain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.kt.String()
			if got != tt.want {
				t.Errorf("KeyType(%d).String() = %q, want %q", tt.kt, got, tt.want)
			}
		})
	}
}

func TestKeyType_Valid(t *testing.T) {
	tests := []struct {
		name string
		kt   KeyType
		want bool
	}{
		{"ip is valid", KeyIP, true},
		{"domain is valid", KeyDomain, true},
		{"undefined is invalid", KeyType(3), false},
		{"max uint8 is invalid", KeyType(math.MaxUint8), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.kt.Valid()
			if got != tt.want {
				t.Errorf("KeyType(%d).Valid() = %v, want %v", tt.kt, got, tt.want)
			}
		})
	}
}

// ---------- Config ----------

func TestDefaultConfig_ReturnsSensibleDefaults(t *testing.T) {
	c := DefaultConfig()

	if c.MergeFlushInterval != MinMergeFlushInterval {
		t.Errorf("MergeFlushInterval = %v, want %v", c.MergeFlushInterval, MinMergeFlushInterval)
	}
	if c.ChanSize != DefaultChanSize {
		t.Errorf("ChanSize = %d, want %d", c.ChanSize, DefaultChanSize)
	}
	if c.BatchSize != DefaultBatchSize {
		t.Errorf("BatchSize = %d, want %d", c.BatchSize, DefaultBatchSize)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid defaults",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "interval too small",
			cfg: Config{
				MergeFlushInterval: 50 * time.Millisecond,
				ChanSize:           100,
				BatchSize:          100,
			},
			wantErr: true,
		},
		{
			name: "interval zero",
			cfg: Config{
				MergeFlushInterval: 0,
				ChanSize:           100,
				BatchSize:          100,
			},
			wantErr: true,
		},
		{
			name: "chan size zero",
			cfg: Config{
				MergeFlushInterval: MinMergeFlushInterval,
				ChanSize:           0,
				BatchSize:          100,
			},
			wantErr: true,
		},
		{
			name: "chan size negative",
			cfg: Config{
				MergeFlushInterval: MinMergeFlushInterval,
				ChanSize:           -5,
				BatchSize:          100,
			},
			wantErr: true,
		},
		{
			name: "batch size zero",
			cfg: Config{
				MergeFlushInterval: MinMergeFlushInterval,
				ChanSize:           100,
				BatchSize:          0,
			},
			wantErr: true,
		},
		{
			name: "batch size negative",
			cfg: Config{
				MergeFlushInterval: MinMergeFlushInterval,
				ChanSize:           100,
				BatchSize:          -1,
			},
			wantErr: true,
		},
		{
			name: "exactly minimum interval",
			cfg: Config{
				MergeFlushInterval: MinMergeFlushInterval,
				ChanSize:           1,
				BatchSize:          1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Normalize_ClampsInvalidValues(t *testing.T) {
	cfg := Config{
		MergeFlushInterval: 0,
		ChanSize:           -10,
		BatchSize:          -1,
	}

	cfg.Normalize()

	if cfg.MergeFlushInterval != MinMergeFlushInterval {
		t.Errorf("MergeFlushInterval after Normalize = %v, want %v", cfg.MergeFlushInterval, MinMergeFlushInterval)
	}
	if cfg.ChanSize != DefaultChanSize {
		t.Errorf("ChanSize after Normalize = %d, want %d", cfg.ChanSize, DefaultChanSize)
	}
	if cfg.BatchSize != DefaultBatchSize {
		t.Errorf("BatchSize after Normalize = %d, want %d", cfg.BatchSize, DefaultBatchSize)
	}
}

func TestConfig_Normalize_LeavesValidValues(t *testing.T) {
	cfg := Config{
		MergeFlushInterval: 500 * time.Millisecond,
		ChanSize:           2048,
		BatchSize:          8192,
	}

	cfg.Normalize()

	if cfg.MergeFlushInterval != 500*time.Millisecond {
		t.Errorf("MergeFlushInterval changed unexpectedly to %v", cfg.MergeFlushInterval)
	}
	if cfg.ChanSize != 2048 {
		t.Errorf("ChanSize changed unexpectedly to %d", cfg.ChanSize)
	}
	if cfg.BatchSize != 8192 {
		t.Errorf("BatchSize changed unexpectedly to %d", cfg.BatchSize)
	}
}

// ---------- ResultSchema ----------

func TestResultSchema_Validate(t *testing.T) {
	tests := []struct {
		name    string
		schema  ResultSchema
		wantErr bool
	}{
		{
			name: "valid schema",
			schema: ResultSchema{
				Name:   "ports",
				Parser: func([]string) (Result, error) { return nil, nil },
			},
			wantErr: false,
		},
		{
			name: "missing name",
			schema: ResultSchema{
				Parser: func([]string) (Result, error) { return nil, nil },
			},
			wantErr: true,
		},
		{
			name: "missing parser",
			schema: ResultSchema{
				Name: "ports",
			},
			wantErr: true,
		},
		{
			name:    "empty schema",
			schema:  ResultSchema{},
			wantErr: true,
		},
		{
			name: "name is whitespace only",
			schema: ResultSchema{
				Name:   "   ",
				Parser: func([]string) (Result, error) { return nil, nil },
			},
			wantErr: false, // Validate only checks empty string, not whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ResultSchema.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ---------- ResultFile ----------

func TestResultFile_SizeString(t *testing.T) {
	gb3 := 3.14 * 1024 * 1024 * 1024
	tests := []struct {
		name      string
		sizeBytes int64
		want      string
	}{
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"500 bytes", 500, "500 B"},
		{"exactly 1 KB", 1024, "1.00 KB"},
		{"1.5 KB", 1536, "1.50 KB"},
		{"exactly 1 MB", 1024 * 1024, "1.00 MB"},
		{"2.5 MB", int64(2.5 * 1024 * 1024), "2.50 MB"},
		{"exactly 1 GB", int64(1024) * 1024 * 1024, "1.00 GB"},
		{"3.14 GB", int64(gb3), "3.14 GB"},
		{"below 1 KB", 1023, "1023 B"},
		{"below 1 MB", int64(1023*1024 + 500), "1023.49 KB"},
		{"below 1 GB", int64(1024)*1024*1024 - 10000, "1023.99 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := ResultFile{SizeBytes: tt.sizeBytes}
			got := f.SizeString()
			if got != tt.want {
				t.Errorf("ResultFile{%d}.SizeString() = %q, want %q", tt.sizeBytes, got, tt.want)
			}
		})
	}
}

func TestResultFile_SizeString_Negative(t *testing.T) {
	f := ResultFile{SizeBytes: -1}
	got := f.SizeString()
	// Negative sizes fall into the default case: "%d B"
	if got != "-1 B" {
		t.Errorf("ResultFile{-1}.SizeString() = %q, want %q", got, "-1 B")
	}
}

// ---------- mockResult ----------

func TestMockResult_ImplementsResult(t *testing.T) {
	var _ Result = newMockResult("1.2.3.4", 0.8, "1.2.3.4", "0.8")
}

func TestMockResult_Key(t *testing.T) {
	r := newMockResult("example.com", 0.5, "example.com", "0.5")
	if r.Key() != "example.com" {
		t.Errorf("Key() = %q, want %q", r.Key(), "example.com")
	}
}

func TestMockResult_Score(t *testing.T) {
	r := newMockResult("1.2.3.4", 0.95, "1.2.3.4", "0.95")
	if r.Score() != 0.95 {
		t.Errorf("Score() = %f, want 0.95", r.Score())
	}
}

func TestMockResult_Equal(t *testing.T) {
	a := newMockResult("1.2.3.4", 0.8, "1.2.3.4", "0.8")
	b := newMockResult("1.2.3.4", 0.8, "1.2.3.4", "0.8")
	c := newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9")
	d := newMockResult("5.6.7.8", 0.8, "5.6.7.8", "0.8")

	if !a.Equal(b) {
		t.Error("equal records should be Equal")
	}
	if a.Equal(c) {
		t.Error("different scores should not be Equal")
	}
	if a.Equal(d) {
		t.Error("different keys should not be Equal")
	}
}

func TestMockResult_ToRecord(t *testing.T) {
	r := newMockResult("1.2.3.4", 0.8, "1.2.3.4", "0.8")
	rec := r.ToRecord()
	if len(rec) != 2 || rec[0] != "1.2.3.4" || rec[1] != "0.8" {
		t.Errorf("ToRecord() = %v, want [1.2.3.4 0.8]", rec)
	}
}
