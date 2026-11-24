package netutil

import (
	"errors"
	"fmt"
	"syscall"
	"testing"
)

// TestIsUnreachable_Errnos covers is unreachable errnos.
func TestIsUnreachable_Errnos(t *testing.T) {
	t.Parallel()

	unreachable := []error{
		syscall.ENETUNREACH,
		syscall.EHOSTUNREACH,
		syscall.EADDRNOTAVAIL,
		syscall.EAFNOSUPPORT,
	}
	for _, err := range unreachable {
		if !IsUnreachable(err) {
			t.Errorf("IsUnreachable(%v) = false, want true", err)
		}
	}

	reachable := []error{
		syscall.ECONNREFUSED,
		syscall.ECONNRESET,
		syscall.EACCES,
		errors.New("not an errno"),
	}
	for _, err := range reachable {
		if IsUnreachable(err) {
			t.Errorf("IsUnreachable(%v) = true, want false", err)
		}
	}
}

// TestIsUnreachable_Wrapped covers is unreachable wrapped.
func TestIsUnreachable_Wrapped(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("dial tcp: %w", syscall.ENETUNREACH)
	if !IsUnreachable(err) {
		t.Fatal("IsUnreachable(wrapped ENETUNREACH) = false, want true")
	}
}
