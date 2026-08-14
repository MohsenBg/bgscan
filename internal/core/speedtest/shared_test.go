package speedtest

import "net/http"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// fakeReadCloser lets tests simulate a response body whose Read (and
// optionally Close) fails or behaves in a controlled way. Shared by
// download_test.go and upload_test.go.
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
