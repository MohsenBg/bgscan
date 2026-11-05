package result

import (
	"time"
)

// replaceFile atomically replaces the destination file with the source file.
func replaceFile(src, dst string) error {
	const (
		maxWait   = 1 * time.Second
		firstWait = 25 * time.Millisecond
		maxStep   = 200 * time.Millisecond
	)
	deadline := time.Now().Add(maxWait)
	wait := firstWait
	for {
		err := moveReplace(src, dst)
		if err == nil || !isTransientLock(err) || time.Now().Add(wait).After(deadline) {
			return err
		}
		time.Sleep(wait)
		wait = min(wait*2, maxStep)
	}
}
