package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	config, err := loadConfig(os.LookupEnv)
	if err != nil {
		return err
	}
	defer clear(config.masterKey)
	defer clear(config.tokenKey)
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()
	return serve(ctx, config)
}
