package main

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"time"

	operatorv1 "github.com/Onurlulardan/harpy/contracts/operator/v1"
)

func runRecordReactivation(
	arguments []string,
	input io.Reader,
	clock func() time.Time,
) error {
	flags := flag.NewFlagSet("record-reactivation", flag.ContinueOnError)
	target := flags.String("target-directory", "", "absolute handoff directory")
	baseURL := flags.String("frontend-base-url", "", "frontend public base URL")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		!filepath.IsAbs(*target) || input == nil || clock == nil {
		return fmt.Errorf("record-reactivation arguments are invalid")
	}
	result, err := operatorv1.ReadIdentityReconcileResult(input)
	if err != nil {
		return err
	}
	handoff, err := buildReactivationHandoff(result, *baseURL, clock())
	if err != nil {
		return err
	}
	document, err := encodeReactivationHandoff(handoff)
	if err != nil {
		return err
	}
	return writeAtomic(reactivationHandoffPath(*target), document)
}
