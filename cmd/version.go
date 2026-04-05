package cmd

import "runtime/debug"

const fallbackVersion = "v0.1.7-dev"

var version = ""

func appVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fallbackVersion
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return fallbackVersion
}
