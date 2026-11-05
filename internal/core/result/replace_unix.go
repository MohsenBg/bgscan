//go:build !windows

package result

import "os"

func moveReplace(src, dst string) error { return os.Rename(src, dst) }
func isTransientLock(error) bool        { return false }

func syncDir(dir string) error {
	df, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = df.Close() }()
	return df.Sync()
}
