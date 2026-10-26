package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"time"
)

type probeResult struct {
	target string
	err    error
}

func probeAll(ctx context.Context, config checkConfig, client *http.Client) []string {
	count := len(config.httpTargets) + len(config.tcpTargets)
	results := make(chan probeResult, count)
	for _, target := range config.httpTargets {
		go func() {
			results <- probeResult{"http " + target, probeHTTP(ctx, client, target)}
		}()
	}
	for _, target := range config.tcpTargets {
		go func() {
			results <- probeResult{"tcp " + target, probeTCP(ctx, config.attemptTimeout(), target)}
		}()
	}
	failures := make([]string, 0, count)
	for range count {
		result := <-results
		if result.err != nil {
			failures = append(failures, result.target+": "+result.err.Error())
		}
	}
	sort.Strings(failures)
	return failures
}

func probeHTTP(ctx context.Context, client *http.Client, target string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("request creation failed: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("status %d", response.StatusCode)
	}
	return nil
}

func probeTCP(ctx context.Context, timeout time.Duration, target string) error {
	connection, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", target)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	_ = connection.Close()
	return nil
}
