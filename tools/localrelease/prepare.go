package main

import (
	"context"
	"fmt"
)

func runPrepare(ctx context.Context, arguments []string) error {
	options, err := parsePrepareOptions(arguments)
	if err != nil {
		return err
	}
	return executePrepare(ctx, options)
}

func executePrepare(context.Context, prepareOptions) error {
	return fmt.Errorf("localrelease prepare no longer emits a migration manifest or cutover receipt")
}
