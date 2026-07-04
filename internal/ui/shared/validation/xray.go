package validation

import (
	"errors"

	"bgscan/internal/core/xray"
)

func ValidateXrayLink(link string) error {
	if link == "" {
		return errors.New("link cannot be empty")
	}
	_, err := xray.ParseLink(link)
	return err
}
