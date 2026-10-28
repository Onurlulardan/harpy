package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
)

func runExact(ctx context.Context, arguments []string, stdin io.Reader) (int, error) {
	options, err := parseExactOptions(arguments)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()
	if _, statErr := os.Lstat(options.needleFile); options.allowAbsent && os.IsNotExist(statErr) {
		return 0, nil
	}
	secret, err := readSecretFile(options.needleFile)
	if err != nil {
		return 0, err
	}
	defer clear(secret)
	verified := 0
	if err = scanEnvironment(secret); err != nil {
		return verified, err
	}
	verified++
	count, err := scanPaths(ctx, secret, options.paths, options.allowedFiles[0])
	verified += count
	if err != nil {
		return verified, err
	}
	if options.stdinLabel != "" {
		if stdin == nil {
			return verified, fmt.Errorf("%s is unavailable", options.stdinLabel)
		}
		found, readErr := streamContains(ctx, stdin, secret, 256*1024*1024)
		if readErr != nil {
			return verified, fmt.Errorf("scan %s: %w", options.stdinLabel, readErr)
		}
		if found {
			return verified, fmt.Errorf("plaintext secret found in %s", options.stdinLabel)
		}
		verified++
	}
	if options.postgresDSNFile != "" {
		if err = scanPostgres(ctx, secret, options.postgresDSNFile); err != nil {
			return verified, err
		}
		verified++
	}
	return verified, nil
}

func scanEnvironment(secret []byte) error {
	for _, entry := range os.Environ() {
		if bytes.Contains([]byte(entry), secret) {
			name := entry
			if index := bytes.IndexByte([]byte(entry), '='); index >= 0 {
				name = entry[:index]
			}
			return fmt.Errorf("plaintext secret found in environment key %s", name)
		}
	}
	return nil
}
