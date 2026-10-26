package main

import (
	"flag"
	"fmt"
	"os"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

type options struct {
	source string
	api    string
	worker string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "recovery-projector:", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	flags := flag.NewFlagSet("recovery-projector", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	config := options{}
	flags.StringVar(&config.source, "source", "", "recovery signing source")
	flags.StringVar(&config.api, "api-target", "", "API recovery target")
	flags.StringVar(&config.worker, "worker-target", "", "worker recovery target")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 {
		return fmt.Errorf("invalid invocation")
	}
	if err := validateOptions(config); err != nil {
		return err
	}
	bundle, controlSeed, err := readSource(config.source)
	if err != nil {
		return err
	}
	defer clear(controlSeed)
	if err = publishExact(config.api, map[string][]byte{
		recoveryv1.BundleFilename:        bundle,
		recoveryv1.ControlSignerFilename: controlSeed,
	}); err != nil {
		return err
	}
	return publishExact(config.worker, map[string][]byte{
		recoveryv1.BundleFilename: bundle,
	})
}
