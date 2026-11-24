package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

// Expected probe failures. NormalizeErr maps low-level errors to
// these values, and the engine counts failures by them.
//
//   - ErrTimeout: the probe passed its deadline.
//   - ErrRefused: the target refused the connection.
//   - ErrUnreachable: the target host could not be reached.
//   - ErrBadResponse: the target answered, but the answer failed the
//     probe checks, such as a rejected status code.
var (
	ErrTimeout     = errors.New("timeout")
	ErrRefused     = errors.New("connection refused")
	ErrUnreachable = errors.New("host unreachable")
	ErrBadResponse = errors.New("bad response")

	// ErrDomainTarget means a probe got a domain row. Probes work on
	// addresses, and nothing resolves names yet. Failing here beats probing
	// an unset address and reporting a misleading dial error.
	ErrDomainTarget = errors.New("domain target needs DNS resolution")
)

// NormalizeErr maps a low-level probe error to one of the shared
// sentinel errors in this package, wrapped around the original error.
// Timeouts, refused connections, and unreachable hosts are mapped.
// Any other error, and nil, is returned unchanged.
func NormalizeErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrRefused) ||
		errors.Is(err, ErrUnreachable) ||
		errors.Is(err, ErrBadResponse) {
		return err
	}

	var ne net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, os.ErrDeadlineExceeded),
		errors.As(err, &ne) && ne.Timeout():
		return fmt.Errorf("%w: %w", ErrTimeout, err)

	case errors.Is(err, syscall.ECONNREFUSED):
		return fmt.Errorf("%w: %w", ErrRefused, err)

	case errors.Is(err, syscall.EHOSTUNREACH),
		errors.Is(err, syscall.ENETUNREACH):
		return fmt.Errorf("%w: %w", ErrUnreachable, err)

	}

	return err
}
