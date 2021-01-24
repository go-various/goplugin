package pluginutil

import (
	"os"

	version "github.com/hashicorp/go-version"
)

var (
	// PluginVersionEnv is the ENV name used to pass the version of the
	// PLUGIN server to the plugins
	PluginVersionEnv = "PLUGIN_VERSION"
)

// GRPCSupport defaults to returning true, unless PLUGIN_VERSION is missing or
// it fails to meet the version constraint.
func GRPCSupport() bool {
	verString := os.Getenv(PluginVersionEnv)
	// If the env var is empty, we fall back to netrpc for backward compatibility.
	if verString == "" {
		return false
	}
	if verString != "unknown" {
		ver, err := version.NewVersion(verString)
		if err != nil {
			return true
		}
		// Due to some regressions on 0.9.2 & 0.9.3 we now require version 0.9.4
		// to allow the plugins framework to default to gRPC.
		constraint, err := version.NewConstraint(">= 0.9.4")
		if err != nil {
			return true
		}
		return constraint.Check(ver)
	}
	return true
}
