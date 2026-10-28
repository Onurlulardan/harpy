package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func parseRefreshSignedReleaseOptions(arguments []string) (string, string, error) {
	flags := flag.NewFlagSet("refresh-signed-release", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	var directory, releaseID string
	flags.StringVar(&directory, "output", "", "marker-owned release output directory")
	flags.StringVar(&releaseID, "release-id", "", "optional applied release id to restore")
	if err := flags.Parse(arguments); err != nil {
		return "", "", err
	}
	if flags.NArg() != 0 {
		return "", "", fmt.Errorf("refresh-signed-release accepts flags only")
	}
	if err := requireAbsolutePath(&directory); err != nil {
		return "", "", err
	}
	return directory, strings.TrimSpace(releaseID), nil
}
