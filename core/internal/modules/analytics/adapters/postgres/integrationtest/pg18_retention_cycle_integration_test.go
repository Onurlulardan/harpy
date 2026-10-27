//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"
)

type analyticsRetentionOutcome struct {
	successorID, payloadReference, payloadSHA256, dedupeKey string
	successorRunAt                                          time.Time
	priority                                                int
	draftsPurged, versionsPurged, releasesPurged            int64
	executionsPurged, auditsPurged, capabilitiesPurged      int64
	legalHoldBlocked, hasRemaining                          bool
	checkpointRevision                                      int64
}

func (outcome analyticsRetentionOutcome) totalPurged() int64 {
	return outcome.draftsPurged + outcome.versionsPurged + outcome.releasesPurged +
		outcome.executionsPurged + outcome.auditsPurged + outcome.capabilitiesPurged
}

func executeAnalyticsRetention(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	jobID, successorID string, limit int,
) analyticsRetentionOutcome {
	t.Helper()
	reference := "analytics-retention/" + analyticsTenantA
	digest := sha256.Sum256([]byte(reference))
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	var outcome analyticsRetentionOutcome
	err = transaction.QueryRowContext(ctx, `SELECT * FROM analytics_execute_retention_cycle_v1(
't_analytics_a'::name,$1::uuid,$2::uuid,$3::uuid,$4::uuid,$5,$6,$7)`,
		analyticsTenantA, jobID,
		"019b1000-0000-7000-8000-000000000118", successorID,
		reference, hex.EncodeToString(digest[:]), limit).Scan(
		&outcome.successorID, &outcome.successorRunAt, &outcome.payloadReference,
		&outcome.payloadSHA256, &outcome.dedupeKey, &outcome.priority,
		&outcome.draftsPurged, &outcome.versionsPurged, &outcome.releasesPurged,
		&outcome.executionsPurged, &outcome.auditsPurged,
		&outcome.capabilitiesPurged, &outcome.legalHoldBlocked,
		&outcome.hasRemaining, &outcome.checkpointRevision)
	if err != nil {
		t.Fatalf("execute Analytics retention: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return outcome
}

func analyticsRetentionRuntimeExec(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	statement string, arguments ...any,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err == nil {
		_, err = transaction.ExecContext(ctx, statement, arguments...)
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("execute Analytics retention authority: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
