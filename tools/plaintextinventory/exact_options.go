package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"
)

type stringList []string

func (values *stringList) String() string { return fmt.Sprint([]string(*values)) }
func (values *stringList) Set(value string) error {
	*values = append(*values, value)
	return nil
}

type exactOptions struct {
	needleFile, postgresDSNFile, stdinLabel string
	paths, allowedFiles                     stringList
	timeout                                 time.Duration
	allowAbsent                             bool
}

func parseExactOptions(arguments []string) (exactOptions, error) {
	flags := flag.NewFlagSet("exact", flag.ContinueOnError)
	var options exactOptions
	flags.StringVar(&options.needleFile, "needle-file", "", "owner-only secret source")
	flags.StringVar(&options.postgresDSNFile, "postgres-dsn-file", "", "PostgreSQL DSN")
	flags.StringVar(&options.stdinLabel, "stdin-label", "", "bounded stdin surface label")
	flags.Var(&options.paths, "path", "file or directory to scan")
	flags.Var(&options.allowedFiles, "allow-file", "single expected secret source")
	flags.DurationVar(&options.timeout, "timeout", 2*time.Minute, "scan timeout")
	flags.BoolVar(&options.allowAbsent, "allow-absent", false, "accept completed cleanup")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		!filepath.IsAbs(options.needleFile) || options.timeout <= 0 ||
		len(options.paths) == 0 || len(options.allowedFiles) != 1 {
		return exactOptions{}, fmt.Errorf("exact scan arguments are invalid")
	}
	for _, path := range append(append(stringList{}, options.paths...), options.allowedFiles...) {
		if !filepath.IsAbs(path) {
			return exactOptions{}, fmt.Errorf("exact scan paths must be absolute")
		}
	}
	if options.postgresDSNFile != "" && !filepath.IsAbs(options.postgresDSNFile) {
		return exactOptions{}, fmt.Errorf("PostgreSQL DSN path must be absolute")
	}
	return options, nil
}
