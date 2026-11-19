package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

type mockResult struct {
	key   string
	score float64
}

func (r *mockResult) Key() string          { return r.key }
func (r *mockResult) KeyType() target.Kind { return target.KindIP }
func (r *mockResult) ToRecord() []string   { return []string{r.key} }
func (r *mockResult) Score() float64       { return r.score }
func (r *mockResult) Equal(o result.Result) bool {
	other, ok := o.(*mockResult)
	return ok && other.key == r.key
}

func mockResultParser(record []string) (result.Result, error) {
	if len(record) != 2 {
		return nil, fmt.Errorf("invalid record length: %d", len(record))
	}
	s, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return nil, err
	}
	return &mockResult{key: record[0], score: s}, nil
}

func mockResultSchema() result.ResultSchema {
	return result.ResultSchema{
		Name:      "mock",
		Directory: "mock",
		Columns:   []result.ColumnDef{{Name: "id", Width: 10}, {Name: "value", Width: 20}},
		Parser:    mockResultParser,
	}
}

type mockProbe struct {
	initErr  error
	closeErr error
	runErr   error
	runDelay time.Duration
	onRun    func(entry target.Entry)

	initCalled  atomic.Int32
	closeCalled atomic.Int32
	runCalled   atomic.Int32
}

func (p *mockProbe) Init(_ context.Context) error { p.initCalled.Add(1); return p.initErr }
func (p *mockProbe) Close() error                 { p.closeCalled.Add(1); return p.closeErr }
func (p *mockProbe) Schema() result.ResultSchema  { return mockResultSchema() }
func (p *mockProbe) Input() target.Kind           { return target.KindIP }
func (p *mockProbe) Run(ctx context.Context, entry target.Entry) (result.Result, error) {
	p.runCalled.Add(1)
	if p.runDelay > 0 {
		select {
		case <-time.After(p.runDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if p.onRun != nil {
		p.onRun(entry)
	}
	if p.runErr != nil {
		return nil, p.runErr
	}
	return &mockResult{key: entry.Value(), score: 1.0}, nil
}

type filteringProbe struct {
	failIPs     map[string]bool
	initCalled  atomic.Int32
	closeCalled atomic.Int32
}

func (p *filteringProbe) Init(_ context.Context) error { p.initCalled.Add(1); return nil }
func (p *filteringProbe) Close() error                 { p.closeCalled.Add(1); return nil }
func (p *filteringProbe) Schema() result.ResultSchema  { return mockResultSchema() }
func (p *filteringProbe) Input() target.Kind           { return target.KindIP }
func (p *filteringProbe) Run(_ context.Context, entry target.Entry) (result.Result, error) {
	ip := entry.IP.Addr().String()
	if p.failIPs[ip] {
		return nil, errors.New("filtered")
	}
	return &mockResult{key: ip}, nil
}

type mockWriter struct {
	startErr error
	stopErr  error
	writeErr error

	mu      sync.Mutex
	written []result.Result

	startCalled atomic.Int32
	stopCalled  atomic.Int32

	resultPath string
}

func (w *mockWriter) Start() error { w.startCalled.Add(1); return w.startErr }
func (w *mockWriter) Close() error { w.stopCalled.Add(1); return w.stopErr }
func (w *mockWriter) GetResultPath() string {
	return w.resultPath
}
func (w *mockWriter) Write(r result.Result) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writeErr != nil {
		return w.writeErr
	}
	w.written = append(w.written, r)
	return nil
}
func (w *mockWriter) Stats() result.WriterStats { return result.WriterStats{} }

func (w *mockWriter) results() []result.Result {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]result.Result, len(w.written))
	copy(out, w.written)
	return out
}

func ipFile(t *testing.T, ips ...string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "ips-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, ip := range ips {
		_, _ = fmt.Fprintf(f, "%s,0,,1\n", ip)
	}
	_ = f.Close()
	return f.Name()
}
