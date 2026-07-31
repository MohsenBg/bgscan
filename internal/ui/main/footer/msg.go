package footer

// UpdateAppVersion requests the footer to display a new app version.
type UpdateAppVersion struct {
	AppVersion string
}

// UpdateStatus requests the footer to display a new status string.
type UpdateStatus struct {
	Status string
}

// NewUpdateAppVersion returns an UpdateAppVersion message.
func NewUpdateAppVersion(version string) UpdateAppVersion {
	return UpdateAppVersion{AppVersion: version}
}

// NewUpdateStatus returns an UpdateStatus message.
func NewUpdateStatus(status string) UpdateStatus {
	return UpdateStatus{Status: status}
}
