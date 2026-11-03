package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	command, optional, ok := parseCommand(os.Args[1:])
	if !ok {
		fail("usage: frontend-manager <start|stop|status> [--if-present]")
	}

	cfg, err := loadConfig(command == "start" && !optional)
	if err != nil {
		fail(err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	switch command {
	case "start":
		if !cfg.CheckoutPresent {
			fmt.Println("frontend-v3: sibling checkout is absent; skipped")
			return
		}
		err = runStart(ctx, cfg)
	case "stop":
		err = runStop(cfg)
	case "status":
		err = runStatus(cfg)
	default:
		err = fmt.Errorf("unknown command %q; use start, stop or status", os.Args[1])
	}
	if err != nil {
		fail(err.Error())
	}
}

func parseCommand(arguments []string) (string, bool, bool) {
	if len(arguments) == 1 {
		return arguments[0], false, true
	}
	if len(arguments) == 2 && arguments[0] == "start" && arguments[1] == "--if-present" {
		return arguments[0], true, true
	}
	return "", false, false
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, "frontend-v3:", message)
	os.Exit(1)
}
