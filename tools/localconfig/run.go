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
	profile := flags.String("profile", string(profileLocal), "local")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *output == "" || flags.NArg() != 0 || *profile != string(profileLocal) {
		fmt.Fprintln(stderr, "usage: localconfig -output <directory> [-reset]")
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
	if err := ensureEnvLocal(); err != nil {
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
