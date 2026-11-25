package footer

// UpdateAppVersion asks the footer to show a new app version.
type UpdateAppVersion struct{ AppVersion string }

// UpdateStatus asks the footer to show a new status text.
type UpdateStatus struct{ Status string }

// NewUpdateAppVersion creates an UpdateAppVersion message.
func NewUpdateAppVersion(version string) UpdateAppVersion {
	return UpdateAppVersion{AppVersion: version}
}

// NewUpdateStatus creates an UpdateStatus message.
func NewUpdateStatus(status string) UpdateStatus {
	return UpdateStatus{Status: status}
}
