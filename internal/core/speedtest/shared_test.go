package speedtest

import (
	"io"
	"net/http"
	"sync"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// fakeReadCloser lets tests simulate a response body whose Read (and
// optionally Close) fails or behaves in a controlled way.
type fakeReadCloser struct {
	readFn  func([]byte) (int, error)
	closeFn func() error
}

func (f fakeReadCloser) Read(p []byte) (int, error) {
	return f.readFn(p)
}

func (f fakeReadCloser) Close() error {
	if f.closeFn != nil {
		return f.closeFn()
	}

	return nil
}

// blockingReadCloser blocks in Read until Close is called. It models a
// stalled response body so tests can verify that a transfer timeout actually
// interrupts the read instead of waiting for the peer.
type blockingReadCloser struct {
	closed chan struct{}
	once   sync.Once
}

func newBlockingReadCloser() *blockingReadCloser {
	return &blockingReadCloser{closed: make(chan struct{})}
}

func (b *blockingReadCloser) Read([]byte) (int, error) {
	<-b.closed
	return 0, io.ErrClosedPipe
}

func (b *blockingReadCloser) Close() error {
	b.once.Do(func() { close(b.closed) })
	return nil
}
