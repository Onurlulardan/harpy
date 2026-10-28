package main

import (
	"flag"
	"fmt"
	"io"
	"time"
)

func runCLI(
	arguments []string,
	stdout io.Writer,
	stderr io.Writer,
	clock func() time.Time,
) int {
	flags := flag.NewFlagSet("localconfig", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("output", "", "local secret output directory")
	reset := flags.Bool("reset", false, "remove an intact tool-owned local secret set")
	profile := flags.String("profile", string(profileLocal), "local or forgeguard")
	publicURL := flags.String("public-url", "", "forgeguard public application URL")
	forgeGuardURL := flags.String("forgeguard-url", "", "forgeguard control-plane URL")
	licenseKind := flags.String("license-kind", "", "connected or airgap")
	inputDirectory := flags.String("input-directory", "", "forgeguard owner input directory")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *output == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: runtime-config -output <directory> [-profile local|forgeguard]")
		return 2
	}
	if *profile == string(profileForgeGuard) {
		if *reset {
			fmt.Fprintln(stderr, "runtime-config: forgeguard reset is forbidden")
			return 2
		}
		options, err := forgeGuardGenerationOptions(
			*publicURL, *forgeGuardURL, *licenseKind, *inputDirectory,
		)
		if err != nil {
			fmt.Fprintf(stderr, "runtime-config: %v\n", err)
			return 2
		}
		return runForgeGuardConfig(*output, options, stdout, stderr, clock)
	}
	if *profile != string(profileLocal) || *publicURL != "" ||
		*forgeGuardURL != "" || *licenseKind != "" || *inputDirectory != "" {
		fmt.Fprintln(stderr, "localconfig: profile options are invalid")
		return 2
	}
	if *reset {
		removed, resolved, err := resetArtifactSet(*output, clock().UTC())
		if err != nil {
			fmt.Fprintf(stderr, "localconfig: %v\n", err)
			return 1
		}
		state := "already absent"
		if removed {
			state = "reset"
		}
		fmt.Fprintf(stdout, "local config %s: %s\n", state, resolved)
		return 0
	}
	result, err := ensureArtifactSet(*output, clock().UTC())
	if err != nil {
		fmt.Fprintf(stderr, "localconfig: %v\n", err)
		return 1
	}
	state := "current"
	if result.Generated {
		state = "generated"
	}
	fmt.Fprintf(stdout, "local config %s: %s (generation %s)\n",
		state, result.Output, result.Manifest.GenerationID)
	return 0
}

func runForgeGuardConfig(
	output string,
	options generationOptions,
	stdout io.Writer,
	stderr io.Writer,
	clock func() time.Time,
) int {
	result, err := ensureForgeGuardArtifactSet(output, clock().UTC(), options)
	if err != nil {
		fmt.Fprintf(stderr, "runtime-config: %v\n", err)
		return 1
	}
	state := "current"
	if result.Generated {
		state = "generated"
	}
	fmt.Fprintf(stdout, "forgeguard runtime config %s: %s (generation %s)\n",
		state, result.Output, result.Manifest.GenerationID)
	return 0
}
