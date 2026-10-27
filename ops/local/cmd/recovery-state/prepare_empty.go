package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	recoverykitv1 "github.com/Onurlulardan/harpy/contracts/recoverykit/v1"
	rehydraterequestv1 "github.com/Onurlulardan/harpy/contracts/rehydraterequest/v1"
	"github.com/google/uuid"
)

func runPrepareEmpty(arguments []string, clock func() time.Time) error {
	flags := flag.NewFlagSet("prepare-empty", flag.ContinueOnError)
	recoverySetID := flags.String("recovery-set-id", "", "recovery set UUID")
	handoff := flags.String("handoff-directory", "/run/handoff", "handoff directory")
	postgresProbe := flags.String("postgres-probe-directory", "/probe/postgres", "database probe")
	policyProbe := flags.String("policy-probe-directory", "/probe/policy", "policy probe")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 {
		return fmt.Errorf("prepare-empty arguments are invalid")
	}
	if !canonicalUUID(*recoverySetID) || !filepath.IsAbs(*handoff) {
		return fmt.Errorf("prepare-empty identity is invalid")
	}
	if err := requireEmptyDirectory(*postgresProbe); err != nil {
		return fmt.Errorf("postgres target is not empty: %w", err)
	}
	if err := requireEmptyDirectory(*policyProbe); err != nil {
		return fmt.Errorf("policy target is not empty: %w", err)
	}
	kitPath := filepath.Join(*handoff, "environment-recovery-kit.v1")
	kitBytes, err := readRegular(kitPath, maximumHandoffBytes)
	if err != nil {
		return err
	}
	kit, err := recoverykitv1.Parse(kitBytes)
	if err != nil {
		return fmt.Errorf("recovery kit is invalid")
	}
	boundNow := clock().UTC().Truncate(time.Microsecond)
	now := boundNow.Format(time.RFC3339Nano)
	probeBytes, err := json.Marshal(emptyProbe{Schema: "empty-target-probe.v1",
		DatabaseAbsent: true, PolicyAbsent: true, ProbedAt: now})
	if err != nil {
		return err
	}
	probeDigest := sha256.Sum256(probeBytes)
	request := rehydraterequestv1.Request{
		RecoverySetID: *recoverySetID, ExpectedEnvironmentID: kit.EnvironmentID(),
		TargetDisposition: "empty", WaiverReason: "empty_no_database_or_policy",
		ProbeSHA256:          hex.EncodeToString(probeDigest[:]),
		DestructiveConfirmed: false, RequestedAt: now}
	requestBytes, err := rehydraterequestv1.Build(request, kitBytes, boundNow)
	if err != nil {
		return err
	}
	if err := writeAtomic(filepath.Join(*handoff, "empty-target-probe.v1.json"), probeBytes); err != nil {
		return err
	}
	return writeAtomic(filepath.Join(*handoff, "rehydrate-request.v1.json"), requestBytes)
}

func requireEmptyDirectory(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("probe path is not absolute")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("probe path is not a real directory")
	}
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	names, err := directory.Readdirnames(1)
	if err != io.EOF || len(names) != 0 {
		return fmt.Errorf("probe directory contains target state")
	}
	return nil
}

func canonicalUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func canonicalNow(clock func() time.Time) string {
	return clock().UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)
}
