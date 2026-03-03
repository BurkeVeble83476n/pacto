package contract

// SupportedSpecVersions lists all known pactoVersion values.
var SupportedSpecVersions = []string{"1.0"}

// IsValidSpecVersion returns true if v is a recognized spec version.
func IsValidSpecVersion(v string) bool {
	for _, sv := range SupportedSpecVersions {
		if sv == v {
			return true
		}
	}
	return false
}
