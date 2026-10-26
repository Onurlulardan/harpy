package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func runResetOperation(arguments []string) error {
	flags := flag.NewFlagSet("reset-operation", flag.ContinueOnError)
	directory := flags.String("state-directory", "/run/restore-state", "restore state")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		!filepath.IsAbs(*directory) {
		return fmt.Errorf("reset-operation arguments are invalid")
	}
	info, err := os.Lstat(*directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("restore state directory is invalid")
	}
	names := []string{
		"progress.json", "drain-ack.json", "receipt.json",
		"security-evidence.json", "iam-key-custody-manifest.json",
		"protected-authority-ack-trust-bundle.v1.json",
	}
	for _, name := range names {
		path := filepath.Join(*directory, name)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("restore state artifact %s is unsafe", name)
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}
