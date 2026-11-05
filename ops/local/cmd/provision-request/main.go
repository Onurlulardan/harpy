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
	if len(arguments) != 0 {
		return errors.New("usage: provision-request")
	}
	return provisionTenant(ctx, processRunner{})
}
