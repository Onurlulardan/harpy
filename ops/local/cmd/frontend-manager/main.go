package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	if len(os.Args) != 2 {
		fail("usage: frontend-manager <start|stop|status>")
	}

	cfg, err := loadConfig()
	if err != nil {
		fail(err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	switch os.Args[1] {
	case "start":
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

func fail(message string) {
	fmt.Fprintln(os.Stderr, "frontend-v3:", message)
	os.Exit(1)
}
