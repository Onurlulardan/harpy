package main

import (
	"context"
	"fmt"
	"io"
)

func run(ctx context.Context, arguments []string, stdin io.Reader, stdout io.Writer) error {
	if ctx == nil || ctx.Err() != nil || len(arguments) == 0 || stdout == nil {
		return fmt.Errorf("invocation is invalid")
	}
	switch arguments[0] {
	case "exact":
		result, err := runExact(ctx, arguments[1:], stdin)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "plaintext inventory: %d surfaces verified\n", result)
		return err
	case "static":
		result, err := runStatic(arguments[1:])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "plaintext inventory: %d source files verified\n", result)
		return err
	default:
		return fmt.Errorf("unknown command")
	}
}
