package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "localrelease:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, arguments []string) error {
	if len(arguments) == 0 {
		return fmt.Errorf("command is required: contract, foundation-contract, foundation-applied, release-applied, prepare, refresh-signed-release, or grant-runtime")
	}
	switch arguments[0] {
	case "contract":
		return runContract(arguments[1:])
	case "foundation-contract":
		return runFoundationContract(arguments[1:])
	case "foundation-applied":
		return runFoundationApplied(ctx, arguments[1:])
	case "release-applied":
		return runReleaseApplied(ctx, arguments[1:])
	case "prepare":
		return runPrepare(ctx, arguments[1:])
	case "refresh-signed-release":
		return runRefreshSignedRelease(arguments[1:])
	case "grant-runtime":
		return runGrantRuntime(ctx, arguments[1:])
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
}
