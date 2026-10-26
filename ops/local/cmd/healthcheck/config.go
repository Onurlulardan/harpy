package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	minimumTimeout = 100 * time.Millisecond
	maximumTimeout = 10 * time.Minute
	maximumAttempt = 2 * time.Second
)

type checkConfig struct {
	httpTargets []string
	tcpTargets  []string
	caFile      string
	timeout     time.Duration
}

func parseConfig(arguments []string) (checkConfig, error) {
	flags := flag.NewFlagSet("healthcheck", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var httpTargets targetFlag
	var tcpTargets targetFlag
	var caFile string
	var timeout time.Duration
	flags.Var(&httpTargets, "http", "HTTP health URL; repeatable")
	flags.Var(&tcpTargets, "tcp", "TCP host:port; repeatable")
	flags.StringVar(&caFile, "ca-file", "", "additional PEM CA file")
	flags.DurationVar(&timeout, "timeout", defaultTimeout, "total retry timeout")
	if err := flags.Parse(arguments); err != nil {
		return checkConfig{}, fmt.Errorf("invalid healthcheck arguments: %w", err)
	}
	if flags.NArg() != 0 {
		return checkConfig{}, errors.New("healthcheck does not accept positional arguments")
	}
	if len(httpTargets)+len(tcpTargets) == 0 {
		return checkConfig{}, errors.New("healthcheck requires at least one -http or -tcp target")
	}
	if timeout < minimumTimeout || timeout > maximumTimeout {
		return checkConfig{}, fmt.Errorf(
			"healthcheck -timeout must be between %s and %s", minimumTimeout, maximumTimeout,
		)
	}
	for _, target := range httpTargets {
		if err := validateHTTPURL(target); err != nil {
			return checkConfig{}, err
		}
	}
	for _, target := range tcpTargets {
		if err := validateTCPAddress(target); err != nil {
			return checkConfig{}, err
		}
	}
	return checkConfig{httpTargets, tcpTargets, caFile, timeout}, nil
}

func validateHTTPURL(target string) error {
	parsed, err := url.ParseRequestURI(target)
	if err != nil || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("invalid -http target %q", target)
	}
	return nil
}

func validateTCPAddress(target string) error {
	host, portValue, err := net.SplitHostPort(target)
	port, portErr := strconv.Atoi(portValue)
	if err != nil || portErr != nil || host == "" || port < 1 || port > 65535 {
		return fmt.Errorf("invalid -tcp target %q", target)
	}
	return nil
}

func (config checkConfig) attemptTimeout() time.Duration {
	if config.timeout < maximumAttempt {
		return config.timeout
	}
	return maximumAttempt
}
