//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/jackc/pgx/v5"
)

func verifyContractRetirementRuntime(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	runtime *ContractRetirementDatabase, authority *ContractRetirementDatabase,
	identity contractretirement.Identity,
) {
	t.Helper()
	active := contractRetirementKey(t, "a")
	other := contractRetirementKey(t, "b")
	for _, key := range []contractretirement.Key{active, other} {
		revision, err := authority.Provision(ctx, key, identity, contractRetirementCount(t, 1))
		if err != nil || revisionValue(t, revision) != 1 {
			t.Fatalf("provision revision = %d err=%v", revisionValue(t, revision), err)
		}
	}
	batch := contractRetirementBatch(t, identity, other, active)
	decision, err := runtime.AdmitManifest(ctx, batch)
	if err != nil {
		t.Fatal(err)
	}
	if count, _ := decision.RetiredCount(); count != 0 {
		t.Fatalf("admission retired count = %d", count)
	}
	decision, err = runtime.RecordHits(ctx, batch)
	if err != nil {
		t.Fatal(err)
	}
	if count, _ := decision.RetiredCount(); count != 0 {
		t.Fatalf("hit retired count = %d", count)
	}
	var hitSum, revisionSum int64
	err = database.admin.QueryRow(ctx, `SELECT sum(hit_count), sum(revision)
FROM platform.contract_retirements WHERE contract_key = ANY($1::text[])`,
		[]string{keyValue(t, active), keyValue(t, other)}).Scan(&hitSum, &revisionSum)
	if err != nil || hitSum != 2 || revisionSum != 4 {
		t.Fatalf("multi-key hit = %d/%d err=%v", hitSum, revisionSum, err)
	}
	missing := contractRetirementBatch(t, identity, contractRetirementKey(t, "c"))
	_, err = runtime.RecordHits(ctx, missing)
	if !contractretirement.FaultIs(err, contractretirement.FaultMissing) {
		t.Fatalf("missing fault = %v", err)
	}
	wrongIdentity, _ := contractretirement.NewIdentity("other", "forms")
	_, err = runtime.RecordHits(ctx, contractRetirementBatch(t, wrongIdentity, active))
	if !contractretirement.FaultIs(err, contractretirement.FaultIdentityMismatch) {
		t.Fatalf("identity fault = %v", err)
	}
	verifyContractRetirementAllOrNone(t, ctx, database, runtime, identity, active)
}

func verifyContractRetirementReadOnlyAdmission(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	connection *pgx.Conn, identity contractretirement.Identity,
) {
	t.Helper()
	key := contractRetirementKey(t, "a")
	value := keyValue(t, key)
	var beforeHit, beforeRevision int64
	if err := database.admin.QueryRow(ctx, `SELECT hit_count,revision
FROM platform.contract_retirements WHERE contract_key=$1`, value).Scan(
		&beforeHit, &beforeRevision,
	); err != nil {
		t.Fatal(err)
	}
	transaction, err := connection.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	adapter, err := NewContractRetirementDatabase(transaction)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := adapter.AdmitManifest(
		ctx, contractRetirementBatch(t, identity, key),
	)
	if err != nil {
		t.Fatalf("read-only admission: %s", err.Error())
	}
	if count, _ := decision.RetiredCount(); count != 0 {
		t.Fatalf("read-only admission retired count = %d", count)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var afterHit, afterRevision int64
	if err := database.admin.QueryRow(ctx, `SELECT hit_count,revision
FROM platform.contract_retirements WHERE contract_key=$1`, value).Scan(
		&afterHit, &afterRevision,
	); err != nil {
		t.Fatal(err)
	}
	if beforeHit != afterHit || beforeRevision != afterRevision {
		t.Fatalf("read-only admission mutated evidence: %d/%d -> %d/%d",
			beforeHit, beforeRevision, afterHit, afterRevision)
	}
}

func verifyContractRetirementAllOrNone(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	runtime *ContractRetirementDatabase, identity contractretirement.Identity,
	active contractretirement.Key,
) {
	t.Helper()
	retired := contractRetirementKey(t, "d")
	_, err := database.admin.Exec(ctx, `INSERT INTO platform.contract_retirements (
contract_key,owner_module,domain,state,remaining_scope_count,hit_count,
zero_hit_since,eligible_at,retired_at,revision,updated_at)
VALUES ($1,'restquery','forms','retired',0,0,clock_timestamp()-interval '31 days',
clock_timestamp()-interval '1 day',clock_timestamp(),1,clock_timestamp())`, keyValue(t, retired))
	if err != nil {
		t.Fatal(err)
	}
	decision, err := runtime.RecordHits(
		ctx, contractRetirementBatch(t, identity, active, retired),
	)
	if err != nil {
		t.Fatal(err)
	}
	isRetired, err := decision.Retires(retired)
	if err != nil || !isRetired {
		t.Fatalf("retired decision = %t err=%v", isRetired, err)
	}
	var hitCount int64
	if err := database.admin.QueryRow(ctx,
		"SELECT hit_count FROM platform.contract_retirements WHERE contract_key=$1",
		keyValue(t, active)).Scan(&hitCount); err != nil || hitCount != 1 {
		t.Fatalf("all-or-none active hit = %d err=%v", hitCount, err)
	}
}
