package version

import "runtime/debug"

// Version is the release version; overridden via -ldflags by goreleaser and
// the build workflow. When unset ("dev"), the binary attempts to recover the
// version from Go module build information (e.g. installed via go install).
var Version = "dev"

// Long returns Version, falling back to the embedded build info when built
// as an untagged module.
func Long() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}
