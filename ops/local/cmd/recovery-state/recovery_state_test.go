package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func TestPrepareEmptyBindsProbeAndExactKitBytes(t *testing.T) {
	root, handoff := t.TempDir(), filepath.Join(t.TempDir(), "handoff")
	postgres, policy := filepath.Join(root, "postgres"), filepath.Join(root, "policy")
	for _, directory := range []string{handoff, postgres, policy} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	kit := recoveryKitFixture(t)
	kitPath := filepath.Join(handoff, "environment-recovery-kit.v1")
	if err := os.WriteFile(kitPath, kit, 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 23, 12, 0, 0, 123456000, time.UTC)
	err := runPrepareEmpty([]string{
		"-recovery-set-id", "018f0000-0000-7000-8000-000000000010",
		"-handoff-directory", handoff,
		"-postgres-probe-directory", postgres,
		"-policy-probe-directory", policy,
	}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	probe := mustReadFile(t, filepath.Join(handoff, "empty-target-probe.v1.json"))
	requestBytes := mustReadFile(t, filepath.Join(handoff, "rehydrate-request.v1.json"))
	var request rehydrateRequest
	if json.Unmarshal(requestBytes, &request) != nil {
		t.Fatal("rehydrate request is invalid JSON")
	}
	probeDigest, kitDigest := sha256.Sum256(probe), sha256.Sum256(kit)
	if request.Schema != "rehydrate-request.v1" ||
		request.ExpectedEnvironmentID != "018f0000-0000-7000-8000-000000000001" ||
		request.TargetDisposition != "empty" ||
		request.WaiverReason != "empty_no_database_or_policy" ||
		request.ProbeSHA256 != hex.EncodeToString(probeDigest[:]) ||
		request.RecoveryKitSHA256 != hex.EncodeToString(kitDigest[:]) ||
		request.DestructiveConfirmed || request.RequestedAt != now.Format(time.RFC3339Nano) {
		t.Fatalf("rehydrate request drifted: %#v", request)
	}
}

func TestPrepareEmptyRejectsAnyExistingTargetState(t *testing.T) {
	root := t.TempDir()
	handoff, postgres, policy := filepath.Join(root, "handoff"),
		filepath.Join(root, "postgres"), filepath.Join(root, "policy")
	for _, directory := range []string{handoff, postgres, policy} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(postgres, "PG_VERSION"), []byte("18"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(handoff, "environment-recovery-kit.v1"),
		recoveryKitFixture(t), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runPrepareEmpty([]string{
		"-recovery-set-id", "018f0000-0000-7000-8000-000000000010",
		"-handoff-directory", handoff,
		"-postgres-probe-directory", postgres,
		"-policy-probe-directory", policy,
	}, time.Now)
	if err == nil {
		t.Fatal("non-empty PostgreSQL target was accepted")
	}
}

func TestAckDrainCopiesExactReadyBinding(t *testing.T) {
	directory := t.TempDir()
	progress := restoreProgress{
		Schema:      "restore-progress.v1",
		OperationID: "018f0000-0000-7000-8000-000000000011",
		AttemptID:   "018f0000-0000-7000-8000-000000000012",
		PlanSHA256:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		State:       "ready_to_stop", UpdatedAt: "2026-07-23T12:00:00Z",
	}
	content, _ := json.Marshal(progress)
	if err := os.WriteFile(filepath.Join(directory, "progress.json"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 23, 12, 1, 0, 0, time.UTC)
	if err := runAckDrain([]string{"-state-directory", directory},
		func() time.Time { return now }); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if err := runAckDrain([]string{"-state-directory", directory},
		func() time.Time { return now }); err != nil {
		t.Fatal("atomic ack replacement failed:", err)
	}
	var ack drainAck
	if json.Unmarshal(mustReadFile(t, filepath.Join(directory, "drain-ack.json")), &ack) != nil {
		t.Fatal("drain ack is invalid JSON")
	}
	if ack.Schema != "restore-drain-ack.v1" ||
		ack.OperationID != progress.OperationID || ack.AttemptID != progress.AttemptID ||
		ack.PlanSHA256 != progress.PlanSHA256 ||
		!reflect.DeepEqual(ack.StoppedServices,
			[]string{"core-api", "core-worker", "frontend-v3"}) ||
		ack.DrainedAt != now.Format(time.RFC3339Nano) {
		t.Fatalf("drain ack drifted: %#v", ack)
	}
}

func TestResetOperationRemovesOnlyKnownArtifacts(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{"progress.json", "receipt.json", "unrelated"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := runResetOperation([]string{"-state-directory", directory}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, "progress.json")); !os.IsNotExist(err) {
		t.Fatal("known progress artifact remains")
	}
	if _, err := os.Stat(filepath.Join(directory, "unrelated")); err != nil {
		t.Fatal("unrelated state was removed")
	}
}

func recoveryKitFixture(t *testing.T) []byte {
	t.Helper()
	content, err := os.ReadFile(
		"../../../../contracts/recoverykit/v1/testdata/browser-vector.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Plaintext json.RawMessage `json:"plaintext"`
	}
	if json.Unmarshal(content, &fixture) != nil {
		t.Fatal("browser vector fixture is invalid")
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(fixture.Plaintext, maximumHandoffBytes)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
