package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"
)

func runInvocation(
	parent context.Context, arguments []string, stdout, stderr io.Writer,
) int {
	flags := flag.NewFlagSet("cutoverauthority", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	configPath := flags.String("config", "", "absolute external authority config path")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 || *configPath == "" {
		_, _ = fmt.Fprintln(stderr, "cutover authority: invalid invocation")
		return 2
	}
	config, err := loadInvocation(*configPath)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cutover authority:", err)
		return 2
	}
	ctx, cancel := context.WithTimeout(
		parent, time.Duration(config.DeadlineSeconds)*time.Second,
	)
	defer cancel()
	var result commandResult
	switch config.Action {
	case "adopt_authority":
		result, err = executeAuthorityAdoption(ctx, config)
	case "issue_finalizer":
		result, err = executeFinalizerIssuance(ctx, config)
	case "authorize_finalization":
		result, err = executeFinalizationAuthorization(ctx, config)
	case "record_full_rollout":
		result, err = executeFullRolloutRecord(ctx, config)
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cutover authority:", err)
		return 10
	}
	if err := writeResult(stdout, result); err != nil {
		_, _ = fmt.Fprintln(stderr, "cutover authority: result publication failed")
		return 10
	}
	return 0
}
