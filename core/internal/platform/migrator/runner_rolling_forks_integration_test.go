//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func assertRollingReleaseForksRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	active *Runner,
) {
	t.Helper()
	activeRecord, err := postgresstore.GetReleaseGeneration(
		ctx, database, rollingRunTwoID,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := postgresstore.CreateReleaseGeneration(
		ctx, database, activeRecord,
	); err != nil {
		t.Fatalf("exact release generation replay failed: %v", err)
	}
	replayFork := activeRecord
	replayFork.CleanupReceiptRunID = "018f0000-0000-7000-8000-000000000399"
	if err := postgresstore.CreateReleaseGeneration(
		ctx, database, replayFork,
	); err == nil || !strings.Contains(err.Error(), "replay differs") {
		t.Fatalf("release generation replay fork error = %v", err)
	}

	sameReleaseFork := *active
	configureRollingRelease(
		&sameReleaseFork, rollingReleaseTwoID, integrationDigest("tampered-release"),
		2, 1, rollingEvidenceTwo(),
	)
	if _, err := sameReleaseFork.Up(
		ctx, UpRequest{RunID: "018f0000-0000-7000-8000-000000000298"},
	); err == nil || !strings.Contains(err.Error(), "tuple differs") {
		t.Fatalf("same release fork error = %v", err)
	}

	generationFork := *active
	generationForkID := "018f0000-0000-7000-8000-000000000301"
	configureRollingRelease(
		&generationFork, generationForkID, integrationDigest("generation-fork"),
		3, 1, rollingEvidence([]string{
			"018f0000-0000-7000-8000-000000000303",
			"018f0000-0000-7000-8000-000000000304",
			"018f0000-0000-7000-8000-000000000305",
		}, byte('3')),
	)
	if _, err := generationFork.Up(
		ctx, UpRequest{RunID: "018f0000-0000-7000-8000-000000000302"},
	); err == nil || !strings.Contains(err.Error(), "database generation manifest fork") {
		t.Fatalf("generation manifest fork error = %v", err)
	}

	duplicateTuple := *active
	duplicateReleaseID := "018f0000-0000-7000-8000-000000000401"
	configureRollingRelease(
		&duplicateTuple, duplicateReleaseID, active.identity.manifestSHA256,
		2, 1, rollingEvidence([]string{
			"018f0000-0000-7000-8000-000000000403",
			"018f0000-0000-7000-8000-000000000404",
			"018f0000-0000-7000-8000-000000000405",
		}, byte('4')),
	)
	if _, err := duplicateTuple.Up(
		ctx, UpRequest{RunID: "018f0000-0000-7000-8000-000000000402"},
	); err == nil || !strings.Contains(err.Error(), "binary tuple is duplicated") {
		t.Fatalf("duplicate binary tuple error = %v", err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := postgresstore.FindReleaseGenerationByRelease(
		ctx, connection, active.identity.environmentID, duplicateReleaseID,
	); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("duplicate release generation persisted: %v", err)
	}
}
