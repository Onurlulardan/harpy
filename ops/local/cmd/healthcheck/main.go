package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := execute(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(arguments []string) error {
	config, err := parseConfig(arguments)
	if err != nil {
		return err
	}
	client, err := newHTTPClient(config.caFile, config.attemptTimeout())
	if err != nil {
		return err
	}
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithTimeout(context.Background(), config.timeout)
	defer cancel()
	return waitForTargets(ctx, config, client)
}
