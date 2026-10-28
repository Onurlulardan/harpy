package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type prepareOptions struct {
	adminDSNFile       string
	runtimePassword    string
	migratorPassword   string
	backupPassword     string
	restorePassword    string
	observerPassword   string
	migrationManifest  string
	outputDirectory    string
	catalogSigningSeed string
	catalogKeyring     string
	replaceOwned       bool
}

type contractOptions struct {
	output, migrationOutput string
}

func parsePrepareOptions(arguments []string) (prepareOptions, error) {
	flags := flag.NewFlagSet("prepare", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	var options prepareOptions
	flags.StringVar(&options.adminDSNFile, "admin-dsn-file", "", "admin DSN secret file")
	flags.StringVar(&options.runtimePassword, "runtime-password-file", "", "runtime password secret file")
	flags.StringVar(&options.migratorPassword, "migrator-password-file", "", "migrator password secret file")
	flags.StringVar(&options.backupPassword, "backup-password-file", "", "backup password secret file")
	flags.StringVar(&options.restorePassword, "restore-password-file", "", "restore password secret file")
	flags.StringVar(
		&options.observerPassword, "reactivation-observer-password-file", "",
		"post-restore reactivation observer password secret file",
	)
	flags.StringVar(&options.migrationManifest, "migration-manifest", "", "canonical migration manifest")
	flags.StringVar(
		&options.catalogSigningSeed, "catalog-package-signing-seed-file", "",
		"Catalog package signing seed secret file",
	)
	flags.StringVar(
		&options.catalogKeyring, "catalog-package-keyring-file", "",
		"Catalog package public keyring file",
	)
	flags.StringVar(&options.outputDirectory, "output", "", "new artifact output directory")
	flags.BoolVar(&options.replaceOwned, "replace-owned-output", false,
		"atomically replace a marker-owned prior release output")
	if err := flags.Parse(arguments); err != nil {
		return prepareOptions{}, err
	}
	if flags.NArg() != 0 {
		return prepareOptions{}, fmt.Errorf("prepare accepts flags only")
	}
	paths := []*string{
		&options.adminDSNFile, &options.runtimePassword,
		&options.migratorPassword, &options.backupPassword, &options.restorePassword,
		&options.observerPassword,
		&options.migrationManifest, &options.catalogSigningSeed,
		&options.catalogKeyring, &options.outputDirectory,
	}
	for _, path := range paths {
		if err := requireAbsolutePath(path); err != nil {
			return prepareOptions{}, err
		}
	}
	return options, nil
}

func parseContractOptions(arguments []string) (contractOptions, error) {
	flags := flag.NewFlagSet("contract", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	var options contractOptions
	flags.StringVar(&options.output, "output", "", "binary schema contract output")
	flags.StringVar(&options.migrationOutput, "migration-output", "", "migration manifest output")
	if err := flags.Parse(arguments); err != nil {
		return contractOptions{}, err
	}
	if flags.NArg() != 0 {
		return contractOptions{}, fmt.Errorf("contract accepts flags only")
	}
	for _, path := range []*string{&options.output, &options.migrationOutput} {
		if err := requireAbsolutePath(path); err != nil {
			return contractOptions{}, err
		}
	}
	return options, nil
}

func parseSinglePath(arguments []string, command, name string) (string, error) {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	var value string
	flags.StringVar(&value, name, "", "required absolute path")
	if err := flags.Parse(arguments); err != nil {
		return "", err
	}
	if flags.NArg() != 0 {
		return "", fmt.Errorf("%s accepts flags only", command)
	}
	if err := requireAbsolutePath(&value); err != nil {
		return "", err
	}
	return value, nil
}

func requireAbsolutePath(value *string) error {
	if *value == "" || !filepath.IsAbs(*value) {
		return fmt.Errorf("all paths must be non-empty absolute paths")
	}
	*value = filepath.Clean(*value)
	return nil
}

func readSecret(path, label string) (string, error) {
	document, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", label, err)
	}
	value := strings.TrimSpace(string(document))
	if value == "" || strings.ContainsAny(value, "\x00\r\n") {
		return "", fmt.Errorf("%s is empty or contains forbidden control bytes", label)
	}
	return value, nil
}
