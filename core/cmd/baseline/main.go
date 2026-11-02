package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/baseline"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "baseline:", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) < 2 {
		return fmt.Errorf("usage: baseline platform <dsn> | baseline tenant <dsn> <schema>")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	switch arguments[0] {
	case "platform":
		if len(arguments) != 2 {
			return fmt.Errorf("usage: baseline platform <dsn>")
		}
		return baseline.ApplyPlatform(ctx, arguments[1])
	case "tenant":
		if len(arguments) != 3 {
			return fmt.Errorf("usage: baseline tenant <dsn> <schema>")
		}
		return baseline.ApplyTenant(ctx, arguments[1], arguments[2])
	default:
		return fmt.Errorf("usage: baseline platform <dsn> | baseline tenant <dsn> <schema>")
	}
}
