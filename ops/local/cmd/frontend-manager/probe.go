package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

func frontendReady(ctx context.Context, readyURL string) bool {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, readyURL, nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 256*1024))
	return err == nil && bytes.Contains(body, []byte(`id="root"`))
}

func waitForFrontend(ctx context.Context, cfg runtimeConfig, command *exec.Cmd) error {
	exited := make(chan error, 1)
	go func() {
		exited <- command.Wait()
	}()

	deadline := time.NewTimer(cfg.StartupWindow)
	defer deadline.Stop()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		if frontendReady(ctx, cfg.ReadyURL) {
			return nil
		}
		select {
		case err := <-exited:
			return fmt.Errorf("yarn dev exited before readiness: %w", err)
		case <-deadline.C:
			return fmt.Errorf("frontend did not become ready within %s", cfg.StartupWindow)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
