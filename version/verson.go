package version

var CurrentCommit string

// BuildVersion is the local build version
const BuildVersion = "0.0.1"

func Version() string {
	return BuildVersion + CurrentCommit
}
