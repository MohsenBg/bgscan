package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

var (
	// Expected probe failures.
	ErrTimeout     = errors.New("timeout")
	ErrRefused     = errors.New("connection refused")
	ErrUnreachable = errors.New("host unreachable")
	ErrBadResponse = errors.New("bad response")
)

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
