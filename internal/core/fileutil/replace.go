package fileutil

import (
	"fmt"
	"time"
)

// ReplaceFile moves src onto dst, replacing dst when it exists.
//
// On Unix this is one rename call. Windows needs MoveFileEx plus a retry loop,
// because a file that is open by Windows defender makes the move fail.
func ReplaceFile(src, dst string) error {
	const (
		maxWait   = 10 * time.Second
		firstWait = 25 * time.Millisecond
		maxStep   = 500 * time.Millisecond
	)

	deadline := time.Now().Add(maxWait)
	wait := firstWait

	for {
		err := moveReplace(src, dst)
		if err == nil {
			return nil
		}
		if !isTransientLock(err) || time.Now().Add(wait).After(deadline) {
			return fmt.Errorf("replace %q with %q: %w", dst, src, err)
		}

		time.Sleep(wait)
		wait = min(wait*2, maxStep)
	}
}
