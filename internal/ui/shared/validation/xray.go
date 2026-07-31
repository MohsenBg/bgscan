package validation

import (
	"errors"

	"bgscan/internal/core/xray"
)

// ValidateXrayLink returns an error if the provided link is empty or cannot be
// parsed as a valid XRay link.
func ValidateXrayLink(link string) error {
	if link == "" {
		return errors.New("link cannot be empty")
	}
	_, err := xray.ParseLink(link)
	return err
}
