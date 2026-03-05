package cli

var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func GetVersion() string {
	return Version
}

func GetFullVersion() map[string]string {
	return map[string]string{
		"version":   Version,
		"gitCommit": GitCommit,
		"buildDate": BuildDate,
	}
}

func SetVersion(version, commit, date string) {
	if version != "" {
		Version = version
	}
	if commit != "" {
		GitCommit = commit
	}
	if date != "" {
		BuildDate = date
	}
}
