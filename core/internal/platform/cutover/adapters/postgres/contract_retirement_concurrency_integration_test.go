//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/jackc/pgx/v5"
)

type contractRetirementRaceResult struct {
	decision contractretirement.Decision
	err      error
}

type contractRetirementProvisionResult struct {
	revision contractretirement.Revision
	err      error
}

func verifyConcurrentContractRetirementProvision(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	identity contractretirement.Identity,
) {
	t.Helper()
	connections := make([]*pgx.Conn, 2)
	adapters := make([]*ContractRetirementDatabase, 2)
	for index := range connections {
		connections[index] = openCutoverMigrator(t, ctx, database.dsn)
		defer connections[index].Close(ctx)
		if _, err := connections[index].Exec(ctx, "SET ROLE harpy_owner"); err != nil {
			t.Fatal(err)
		}
		var err error
		adapters[index], err = NewContractRetirementDatabase(connections[index])
		if err != nil {
			t.Fatal(err)
		}
	}
	key := contractRetirementKey(t, "7")
	count := contractRetirementCount(t, 1)
	results := make(chan contractRetirementProvisionResult, 2)
	for _, adapter := range adapters {
		go func(authority *ContractRetirementDatabase) {
			revision, err := authority.Provision(ctx, key, identity, count)
			results <- contractRetirementProvisionResult{revision: revision, err: err}
		}(adapter)
	}
	for range adapters {
		result := <-results
		if result.err != nil {
			t.Fatalf("concurrent provision: %s", result.err.Error())
		}
		if revisionValue(t, result.revision) != 1 {
			t.Fatalf("concurrent provision revision = %d", revisionValue(t, result.revision))
		}
	}
	var rows int
	if err := database.admin.QueryRow(ctx, `SELECT count(*)
FROM platform.contract_retirements WHERE contract_key=$1`, keyValue(t, key)).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("concurrent provision rows = %d", rows)
	}
}

func verifyContractRetirementConcurrency(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	runtime *ContractRetirementDatabase, authority *ContractRetirementDatabase,
	identity contractretirement.Identity,
) {
	t.Helper()
	key := contractRetirementKey(t, "8")
	value := keyValue(t, key)
	_, err := database.admin.Exec(ctx, `INSERT INTO platform.contract_retirements (
contract_key,owner_module,domain,state,remaining_scope_count,hit_count,
zero_hit_since,eligible_at,revision,updated_at)
VALUES ($1,'restquery','analytics','eligible',0,0,clock_timestamp()-interval '31 days',
clock_timestamp(),1,clock_timestamp())`, value)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := database.admin.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback(ctx)
	if _, err := lock.Exec(ctx, `SELECT contract_key
FROM platform.contract_retirements WHERE contract_key=$1 FOR UPDATE`, value); err != nil {
		t.Fatal(err)
	}
	hitResult := make(chan contractRetirementRaceResult, 1)
	finalizeResult := make(chan error, 1)
	batch := contractRetirementBatch(t, identity, key)
	revision, _ := contractretirement.NewRevision(1)
	go func() {
		decision, err := runtime.RecordHits(ctx, batch)
		hitResult <- contractRetirementRaceResult{decision: decision, err: err}
	}()
	go func() {
		_, err := authority.Finalize(ctx, key, identity, revision)
		finalizeResult <- err
	}()
	waitForContractRetirementLocks(t, ctx, database.admin)
	if err := lock.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	hit := <-hitResult
	finalizeErr := <-finalizeResult
	if hit.err != nil {
		t.Fatalf("concurrent hit: %s", hit.err.Error())
	}
	retired, err := hit.decision.Retires(key)
	if err != nil {
		t.Fatal(err)
	}
	if finalizeErr == nil {
		if !retired {
			t.Fatal("finalize won but hit did not return retired decision")
		}
	} else if retired || (!contractretirement.FaultIs(
		finalizeErr, contractretirement.FaultRevisionMismatch,
	) && !contractretirement.FaultIs(finalizeErr, contractretirement.FaultNotEligible)) {
		t.Fatalf("hit/finalize race result is invalid: retired=%t fault=%s",
			retired, contractRetirementFaultName(finalizeErr))
	}
	assertContractRetirementRaceState(t, ctx, database.admin, value, retired)
}

func waitForContractRetirementLocks(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting int
		err := admin.QueryRow(ctx, `SELECT count(*) FROM pg_catalog.pg_stat_activity
WHERE datname=current_database() AND wait_event_type='Lock'
AND (query LIKE '%contract_retirement_hit_v1%'
     OR query LIKE '%contract_retirement_finalize_v1%')`).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting == 2 {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("waiting retirement operations = %d, want 2", waiting)
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}

func assertContractRetirementRaceState(
	t *testing.T, ctx context.Context, admin *pgx.Conn, key string, retired bool,
) {
	t.Helper()
	var state string
	var hitCount, revision int64
	if err := admin.QueryRow(ctx, `SELECT state,hit_count,revision
FROM platform.contract_retirements WHERE contract_key=$1`, key).Scan(
		&state, &hitCount, &revision,
	); err != nil {
		t.Fatal(err)
	}
	if retired && (state != "retired" || hitCount != 0 || revision != 2) {
		t.Fatalf("finalize-win row = %s/%d/%d", state, hitCount, revision)
	}
	if !retired && (state != "active" || hitCount != 1 || revision != 2) {
		t.Fatalf("hit-win row = %s/%d/%d", state, hitCount, revision)
	}
}
