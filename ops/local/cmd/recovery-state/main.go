package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], time.Now))
}

func run(arguments []string, clock func() time.Time) int {
	if len(arguments) == 0 {
		return 2
	}
	var err error
	switch arguments[0] {
	case "stage-plan":
		err = runStage(arguments[1:], "restore-plan.v1.json", false)
	case "stage-kit":
		err = runStage(arguments[1:], "environment-recovery-kit.v1", true)
	case "prepare-empty":
		err = runPrepareEmpty(arguments[1:], clock)
	case "wait-ready":
		err = runWaitReady(arguments[1:])
	case "ack-drain":
		err = runAckDrain(arguments[1:], clock)
	case "verify-completed":
		err = runVerifyCompleted(arguments[1:])
	case "reset-operation":
		err = runResetOperation(arguments[1:])
	case "record-reactivation":
		err = runRecordReactivation(arguments[1:], os.Stdin, clock)
	case "show-reactivation":
		err = runShowReactivation(arguments[1:], os.Stdout)
	default:
		return 2
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "local recovery:", err)
		return 1
	}
	return 0
}

func parseStageFlags(arguments []string) (string, string, error) {
	flags := flag.NewFlagSet("stage", flag.ContinueOnError)
	source := flags.String("source", "", "absolute source artifact")
	target := flags.String("target-directory", "", "absolute handoff directory")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		*source == "" || *target == "" {
		return "", "", fmt.Errorf("stage arguments are invalid")
	}
	return *source, *target, nil
}
