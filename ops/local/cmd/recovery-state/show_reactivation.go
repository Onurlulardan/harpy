package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path/filepath"
)

func runShowReactivation(arguments []string, output io.Writer) error {
	flags := flag.NewFlagSet("show-reactivation", flag.ContinueOnError)
	source := flags.String("source", "", "absolute handoff artifact")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		!filepath.IsAbs(*source) || output == nil {
		return fmt.Errorf("show-reactivation arguments are invalid")
	}
	document, err := readRegular(*source, 64*1024)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var handoff reactivationHandoff
	if decoder.Decode(&handoff) != nil || requireJSONEOF(decoder) != nil || !handoff.valid() {
		return fmt.Errorf("reactivation handoff is invalid")
	}
	_, err = fmt.Fprintf(output, "Post-restore owner reactivation: %s\n", handoff.URL)
	return err
}
