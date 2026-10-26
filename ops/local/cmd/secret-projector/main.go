package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr, time.Now, os.Chown))
}

func run(
	arguments []string,
	stderr io.Writer,
	clock func() time.Time,
	chown ownershipChanger,
) int {
	flags := flag.NewFlagSet("local-secret-projector", flag.ContinueOnError)
	flags.SetOutput(stderr)
	source := flags.String("source", "/source", "verified localconfig source")
	targets := flags.String("targets", "/targets", "mounted target volume root")
	profile := flags.String(
		"profile", "local", "local",
	)
	licenseKind := flags.String("license-kind", "", "connected or airgap")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 {
		return 2
	}
	targetCatalog, stateCatalog, err := projectionCatalog(*profile, *licenseKind)
	if err != nil {
		fmt.Fprintf(stderr, "runtime-secret-projector: %v\n", err)
		return 2
	}
	verified, err := verifySourceFor(
		*source, clock().UTC(), targetCatalog,
	)
	if err != nil {
		fmt.Fprintf(stderr, "runtime-secret-projector: %v\n", err)
		return 1
	}
	for _, target := range targetCatalog {
		path := filepath.Join(*targets, target.name)
		if err := projectTarget(path, verified, target, chown); err != nil {
			fmt.Fprintf(stderr, "runtime-secret-projector: %v\n", err)
			return 1
		}
	}
	for _, state := range stateCatalog {
		path := filepath.Join(*targets, state.name)
		if err := initializeStateVolume(path, state, chown); err != nil {
			fmt.Fprintf(stderr, "runtime-secret-projector: %v\n", err)
			return 1
		}
	}
	return 0
}

func projectionCatalog(
	profile string,
	licenseKind string,
) ([]targetSpec, []targetSpec, error) {
	switch profile {
	case "local":
		if licenseKind != "" {
			return nil, nil, fmt.Errorf("local license kind is forbidden")
		}
		return secretTargets(), stateTargets(), nil
	default:
		return nil, nil, fmt.Errorf("projection profile is invalid")
	}
}
