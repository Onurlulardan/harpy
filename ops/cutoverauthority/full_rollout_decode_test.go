package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	corecutover "github.com/Onurlulardan/harpy/core/public/cutover"
)

func TestFullRolloutEvidenceIsCanonicalAndArtifactBound(t *testing.T) {
	directory := t.TempDir()
	document, artifact := fullRolloutFixture(t)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(directory, "completion.json")
	artifactPath := filepath.Join(directory, "evidence.bin")
	if os.WriteFile(evidencePath, raw, 0o600) != nil ||
		os.WriteFile(artifactPath, artifact, 0o600) != nil {
		t.Fatal("write full rollout fixture")
	}
	completion, err := loadFullRolloutCompletion(evidencePath)
	if err != nil || corecutover.EvaluateFullRollout(completion.Evidence) != nil {
		t.Fatalf("load completion: %v", err)
	}
	if err := verifyFullRolloutArtifacts(
		[]string{artifactPath}, completion.Evidence,
	); err != nil {
		t.Fatal(err)
	}
	if err := verifyFullRolloutArtifacts(nil, completion.Evidence); err == nil {
		t.Fatal("missing evidence artifacts accepted")
	}
	if err := os.WriteFile(evidencePath, append(raw, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadFullRolloutCompletion(evidencePath); err == nil {
		t.Fatal("noncanonical completion evidence accepted")
	}
}

func fullRolloutFixture(t *testing.T) (fullRolloutDocument, []byte) {
	t.Helper()
	artifact := []byte("verified rollout evidence")
	hash := sha256.Sum256(artifact)
	digest := hex.EncodeToString(hash[:])
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	environment := "30000000-0000-4000-8000-000000000001"
	tenant := "40000000-0000-4000-8000-000000000001"
	steps := make([]fullRolloutStep, 0, len(corecutover.RequiredFullRolloutSteps()))
	for index, step := range corecutover.RequiredFullRolloutSteps() {
		scopeID := tenant
		if step.Scope == corecutover.FullRolloutScopeGlobal {
			scopeID = environment
		}
		steps = append(steps, fullRolloutStep{Scope: string(step.Scope),
			Domain: step.Domain, ScopeID: scopeID,
			RequestedBy: "50000000-0000-4000-8000-000000000001",
			ApprovedBy:  "60000000-0000-4000-8000-000000000001",
			Generation:  int64(index + 1), State: "v2", Writer: "v2",
			AcceptanceSHA256: digest, ReverseRehearsalSHA256: digest,
			SourceCount: 10, TargetCount: 10, SourceSHA256: digest,
			TargetSHA256: digest})
	}
	cursor := now.Add(-19 * 24 * time.Hour)
	cohorts := make([]fullRolloutCohort, 0, len(corecutover.FullRolloutCohortOrder()))
	for _, cohort := range corecutover.FullRolloutCohortOrder() {
		minimum, err := corecutover.FullRolloutMinimumSoak(cohort)
		if err != nil {
			t.Fatal(err)
		}
		endedAt := cursor.Add(minimum)
		cohorts = append(cohorts, fullRolloutCohort{Cohort: string(cohort),
			StartedAt:        cursor.Format(time.RFC3339Nano),
			EndedAt:          endedAt.Format(time.RFC3339Nano),
			EligibleRequests: 10_000, ComparedRequests: 10_000})
		cursor = endedAt
	}
	return fullRolloutDocument{ContractVersion: fullRolloutEvidenceContract,
		ReceiptID:     "70000000-0000-4000-8000-000000000001",
		EnvironmentID: environment, RecordedAt: now.Format(time.RFC3339Nano),
		Steps: steps, Cohorts: cohorts, Now: now.Format(time.RFC3339Nano),
		LastLegacyHitAt:  now.Add(-30 * 24 * time.Hour).Format(time.RFC3339Nano),
		ShadowReadSHA256: digest, LogicalDeltaSHA256: digest,
		ReconcileSHA256: digest, ProductionSnapshotSHA256: digest,
		V1RemovalSHA256: digest, BackupRestoreSHA256: digest,
		FrontendContractSHA256: digest, MCPContractSHA256: digest,
		RollbackClosureSHA256:  digest,
		RollbackWindowClosedAt: now.Add(-12 * time.Hour).Format(time.RFC3339Nano),
	}, artifact
}
