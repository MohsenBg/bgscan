package footer

type UpdateAppVersion struct{ AppVersion string }
type UpdateStatus struct{ Status string }

func NewUpdateAppVersion(version string) UpdateAppVersion {
	return UpdateAppVersion{AppVersion: version}
}

func NewUpdateStatus(status string) UpdateStatus {
	return UpdateStatus{Status: status}
}
