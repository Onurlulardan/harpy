package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const retryInterval = 200 * time.Millisecond

func waitForTargets(
	ctx context.Context,
	config checkConfig,
	client *http.Client,
) error {
	var failures []string
	for {
		failures = probeAll(ctx, config, client)
		if len(failures) == 0 {
			return nil
		}
		timer := time.NewTimer(retryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf(
				"healthcheck timed out after %s: %s",
				config.timeout,
				strings.Join(failures, "; "),
			)
		case <-timer.C:
		}
	}
}
