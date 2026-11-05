package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func main() {
	if err := execute(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func execute(ctx context.Context, arguments []string) error {
	switch {
	case len(arguments) == 0:
		return provisionTenant(ctx, processRunner{})
	case len(arguments) == 1 && arguments[0] == "wait-operator":
		return waitOperator(ctx, processRunner{})
	default:
		return errors.New("usage: provision-request [wait-operator]")
	}
}
