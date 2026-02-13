package version

// These variables are intended to be set at build time via -ldflags.
// Defaults keep local development simple.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	BuiltBy string `json:"builtBy"`
}

func Info() BuildInfo {
	return BuildInfo{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
		BuiltBy: BuiltBy,
	}
}
