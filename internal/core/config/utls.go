package config

import (
	"strings"

	utls "github.com/refraction-networking/utls"
)

// utlsClientHelloIDMap maps human-readable fingerprint labels to uTLS ClientHelloIDs.
var utlsClientHelloIDMap = []struct {
	Label string
	ID    *utls.ClientHelloID
}{
	{"random", &utls.HelloRandomizedALPN},
	{"Firefox", &utls.HelloFirefox_Auto},
	{"Firefox_55", &utls.HelloFirefox_55},
	{"Firefox_56", &utls.HelloFirefox_56},
	{"Firefox_63", &utls.HelloFirefox_63},
	{"Firefox_65", &utls.HelloFirefox_65},
	{"Firefox_99", &utls.HelloFirefox_99},
	{"Firefox_102", &utls.HelloFirefox_102},
	{"Firefox_105", &utls.HelloFirefox_105},
	{"Firefox_120", &utls.HelloFirefox_120},
	{"Chrome", &utls.HelloChrome_Auto},
	{"Chrome_58", &utls.HelloChrome_58},
	{"Chrome_62", &utls.HelloChrome_62},
	{"Chrome_70", &utls.HelloChrome_70},
	{"Chrome_72", &utls.HelloChrome_72},
	{"Chrome_83", &utls.HelloChrome_83},
	{"Chrome_87", &utls.HelloChrome_87},
	{"Chrome_96", &utls.HelloChrome_96},
	{"Chrome_100", &utls.HelloChrome_100},
	{"Chrome_102", &utls.HelloChrome_102},
	{"Chrome_120", &utls.HelloChrome_120},
	{"iOS", &utls.HelloIOS_Auto},
	{"iOS_11_1", &utls.HelloIOS_11_1},
	{"iOS_12_1", &utls.HelloIOS_12_1},
	{"iOS_13", &utls.HelloIOS_13},
	{"iOS_14", &utls.HelloIOS_14},
}

// utlsLabels returns all valid fingerprint labels for use in validation and UI.
func utlsLabels() []string {
	labels := make([]string, len(utlsClientHelloIDMap))
	for i, entry := range utlsClientHelloIDMap {
		labels[i] = entry.Label
	}
	return labels
}

// UTLSLookup returns a *utls.ClientHelloID by case-insensitive label match,
// or nil if there is no match.
func UTLSLookup(label string) *utls.ClientHelloID {
	for _, entry := range utlsClientHelloIDMap {
		if strings.EqualFold(label, entry.Label) {
			return entry.ID
		}
	}
	return nil
}

// ValidFingerprint reports whether label is a known uTLS fingerprint.
func ValidFingerprint(label string) bool {
	return UTLSLookup(label) != nil
}

// FingerprintLabels returns the list of valid fingerprint labels.
func FingerprintLabels() []string {
	return utlsLabels()
}
