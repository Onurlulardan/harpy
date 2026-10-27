package buildinfo

import (
	"fmt"
	"runtime"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func Read(service runtimev1.Service, environment runtimev1.Environment) (Info, error) {
	return read(service, environment, releaseVersion, commit, builtAt, runtime.Version())
}

func read(
	service runtimev1.Service,
	environment runtimev1.Environment,
	version string,
	revision string,
	buildTime string,
	goVersion string,
) (Info, error) {
	if goVersion != runtimev1.GoVersion {
		return Info{}, fmt.Errorf("runtime Go version does not match the pinned contract")
	}
	wireVersion, err := runtimev1.NewVersion(service, version, revision, buildTime, environment)
	if err != nil {
		return Info{}, fmt.Errorf("build provenance is invalid: %w", err)
	}
	return Info{version: wireVersion}, nil
}
