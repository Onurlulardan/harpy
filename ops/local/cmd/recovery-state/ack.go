package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func runAckDrain(arguments []string, clock func() time.Time) error {
	flags := flag.NewFlagSet("ack-drain", flag.ContinueOnError)
	directory := flags.String("state-directory", "/run/restore-state", "restore state")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		!filepath.IsAbs(*directory) {
		return fmt.Errorf("ack-drain arguments are invalid")
	}
	progress, err := readProgress(filepath.Join(*directory, "progress.json"))
	if err != nil || progress.State != "ready_to_stop" {
		return fmt.Errorf("restore is not ready_to_stop")
	}
	ack := drainAck{Schema: "restore-drain-ack.v1",
		OperationID: progress.OperationID, AttemptID: progress.AttemptID,
		PlanSHA256:      progress.PlanSHA256,
		StoppedServices: []string{"core-api", "core-worker", "frontend-v3"},
		DrainedAt:       canonicalNow(clock)}
	content, err := json.Marshal(ack)
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(*directory, "drain-ack.json"), content)
}

func runVerifyCompleted(arguments []string) error {
	flags := flag.NewFlagSet("verify-completed", flag.ContinueOnError)
	directory := flags.String("state-directory", "/run/restore-state", "restore state")
	trustPath := flags.String("trust-bundle",
		"/run/recovery-signing/current/recovery-trust-bundle.v1.json", "trust bundle")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		!filepath.IsAbs(*directory) || !filepath.IsAbs(*trustPath) {
		return fmt.Errorf("verify-completed arguments are invalid")
	}
	progress, err := readProgress(filepath.Join(*directory, "progress.json"))
	if err != nil || progress.State != "completed" {
		return fmt.Errorf("restore is not completed")
	}
	bundleBytes, err := readRegular(*trustPath, 2*1024*1024)
	if err != nil {
		return err
	}
	bundle, err := recoverysigningv1.Parse(bundleBytes)
	if err != nil {
		return fmt.Errorf("recovery trust bundle is invalid")
	}
	receiptBytes, err := readRegular(filepath.Join(*directory, "receipt.json"), 2*1024*1024)
	if err != nil {
		return err
	}
	receipt, err := environmentrecoveryv1.Parse(receiptBytes, bundle, time.Now().UTC())
	if err != nil || receipt.Purpose() != environmentrecoveryv1.PurposeRestoreReceipt ||
		receipt.OperationID() != progress.OperationID {
		return fmt.Errorf("restore receipt is invalid")
	}
	return nil
}
