package version

var CurrentCommit string

// BuildVersion is the local build version
const BuildVersion = "v0.0.3"

func Version() string {
	return BuildVersion + CurrentCommit
}
