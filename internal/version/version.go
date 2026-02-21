package version

// These variables are intended to be set at build time via -ldflags.
// Defaults keep local development simple.
var (
	SchemaVersion = "1"
	Version       = "dev"
	Commit        = "none"
	Date          = "unknown"
	BuiltBy       = "unknown"
)

type BuildInfo struct {
	SchemaVersion string `json:"schemaVersion"`
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	Date          string `json:"date"`
	BuiltBy       string `json:"builtBy"`
}

func Info() BuildInfo {
	return BuildInfo{
		SchemaVersion: SchemaVersion,
		Version:       Version,
		Commit:        Commit,
		Date:          Date,
		BuiltBy:       BuiltBy,
	}
}
