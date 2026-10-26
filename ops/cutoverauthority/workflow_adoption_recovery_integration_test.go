//go:build integration

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestAdoptionFailureRestoresExactMigratorStatePG184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startWorkflowDatabase(t, ctx)
	directory := t.TempDir()
	fixture := prepareWorkflowAuthorityArtifacts(
		t, ctx, database, directory, time.Now().UTC().Truncate(time.Second),
	)
	bundle, err := loadAuthorityBundle(fixture.paths, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	policy := bundle.operation.Value().QuiescencePolicy.Roles[0]
	before, err := readQuiescedRole(ctx, database.admin, policy.Principal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.admin.Exec(ctx,
		"ALTER TABLE cutover_authority.finalization_receipts OWNER TO postgres",
	); err != nil {
		t.Fatal(err)
	}
	_, err = executeAuthorityAdoption(ctx, fixture.adoptionInvocation())
	if err == nil || err.Error() == "authority adoption recovery required" {
		t.Fatalf("expected recoverable adoption failure, got %v", err)
	}
	after, observeErr := readQuiescedRole(ctx, database.admin, policy.Principal)
	if observeErr != nil || after.state != policy.PostState ||
		after.verifier != before.verifier {
		t.Fatalf("migrator recovery differs: before=%+v after=%+v err=%v",
			before.state, after.state, observeErr)
	}
}

func TestAdoptionConnectionLossReportsRecoveryRequiredPG184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startWorkflowDatabase(t, ctx)
	directory := t.TempDir()
	fixture := prepareWorkflowAuthorityArtifacts(
		t, ctx, database, directory, time.Now().UTC().Truncate(time.Second),
	)
	bundle, err := loadAuthorityBundle(fixture.paths, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	policy := bundle.operation.Value().QuiescencePolicy.Roles[0]
	blocker, err := pgx.Connect(ctx, database.dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Close(context.Background())
	blockerTx, err := blocker.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blockerTx.Rollback(context.Background())
	if _, err := blockerTx.Exec(ctx,
		"LOCK TABLE platform.cutover_operations IN ACCESS EXCLUSIVE MODE",
	); err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		result commandResult
		err    error
	}
	completed := make(chan outcome, 1)
	go func() {
		result, executeErr := executeAuthorityAdoption(ctx, fixture.adoptionInvocation())
		completed <- outcome{result: result, err: executeErr}
	}()
	pid := waitForQuiescedAuthorityPID(t, ctx, database)
	var terminated bool
	if err := database.admin.QueryRow(ctx,
		"SELECT pg_catalog.pg_terminate_backend($1)", pid,
	).Scan(&terminated); err != nil || !terminated {
		t.Fatalf("terminate authority session: terminated=%v err=%v", terminated, err)
	}
	_ = blockerTx.Rollback(ctx)
	select {
	case value := <-completed:
		if value.err == nil || value.err.Error() != "authority adoption recovery required" ||
			value.result != (commandResult{}) {
			t.Fatalf("connection loss outcome differs: result=%+v err=%v",
				value.result, value.err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("authority connection loss did not return")
	}
	temporary, err := readQuiescedRole(ctx, database.admin, policy.Principal)
	if err != nil || temporary.state != policy.TemporaryState {
		t.Fatalf("expected explicit operator recovery state: %+v err=%v",
			temporary.state, err)
	}
	if err := restoreMigrator(ctx, database.admin, policy); err != nil {
		t.Fatalf("operator recovery failed: %v", err)
	}
}

func waitForQuiescedAuthorityPID(
	t *testing.T, ctx context.Context, database workflowDatabase,
) int32 {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var pid int32
		err := database.admin.QueryRow(ctx, `SELECT activity.pid
FROM pg_catalog.pg_stat_activity AS activity
JOIN pg_catalog.pg_roles AS role ON role.rolname='harpy_migrator'
WHERE activity.application_name=$1 AND activity.usename='postgres'
  AND NOT role.rolcanlogin LIMIT 1`, authorityApplicationName).Scan(&pid)
		if err == nil {
			return pid
		}
		if err != pgx.ErrNoRows && !strings.Contains(err.Error(), "no rows") {
			t.Fatalf("observe authority session: %v", err)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("authority session did not reach post-quiescence lock")
	return 0
}
