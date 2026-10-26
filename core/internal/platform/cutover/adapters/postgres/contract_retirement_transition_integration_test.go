//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/jackc/pgx/v5/pgconn"
)

func verifyContractRetirementTransitions(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	runtime *ContractRetirementDatabase, authority *ContractRetirementDatabase,
	identity contractretirement.Identity,
) {
	t.Helper()
	window := contractRetirementKey(t, "e")
	revision, err := authority.Provision(ctx, window, identity, contractRetirementCount(t, 0))
	if err != nil {
		t.Fatalf("provision zero-scope: %s", err.Error())
	}
	_, err = runtime.AdmitManifest(ctx, contractRetirementBatch(t, identity, window))
	if !contractretirement.FaultIs(err, contractretirement.FaultNotEligible) {
		t.Fatalf("zero-scope admission fault = %v", err)
	}
	_, err = authority.Provision(ctx, window, identity, contractRetirementCount(t, 1))
	if !contractretirement.FaultIs(err, contractretirement.FaultRevisionMismatch) {
		t.Fatalf("unreconciled provision fault = %v", err)
	}
	revision, err = authority.Reconcile(
		ctx, window, identity, contractRetirementCount(t, 1), revision,
	)
	if err != nil {
		var count, storedRevision int64
		var state string
		queryErr := database.admin.QueryRow(ctx, `SELECT remaining_scope_count,state,revision
FROM platform.contract_retirements WHERE contract_key=$1`, keyValue(t, window)).Scan(
			&count, &state, &storedRevision,
		)
		diagnostic := openCutoverMigrator(t, ctx, database.dsn)
		defer diagnostic.Close(ctx)
		_, _ = diagnostic.Exec(ctx, "SET ROLE harpy_owner")
		var ignored int64
		diagnosticErr := diagnostic.QueryRow(ctx, reconcileContractRetirementSQL,
			keyValue(t, window), "restquery", "forms", 1, storedRevision).Scan(&ignored)
		var postgresError *pgconn.PgError
		code, message := "", ""
		if errors.As(diagnosticErr, &postgresError) {
			code, message = postgresError.Code, postgresError.Message
		}
		t.Fatalf("reconcile fault=%s row=%d/%s/%d query=%v db=%s/%s",
			contractRetirementFaultName(err), count, state, storedRevision,
			queryErr, code, message)
	}
	if _, err := runtime.AdmitManifest(
		ctx, contractRetirementBatch(t, identity, window),
	); err != nil {
		t.Fatalf("reconciled admission: %v", err)
	}
	revision, err = authority.Reconcile(
		ctx, window, identity, contractRetirementCount(t, 0), revision,
	)
	if err != nil {
		t.Fatalf("reconcile zero-scope: %s", err.Error())
	}
	verifyContractRetirementBoundary(t, ctx, database, authority, identity, window, revision)
	verifyEligibleHitReset(t, ctx, database, runtime, identity)
	verifyContractRetirementOverflow(t, ctx, database, runtime, identity)
}

func contractRetirementFaultName(err error) string {
	for _, candidate := range []struct {
		kind contractretirement.FaultKind
		name string
	}{
		{contractretirement.FaultDependency, "dependency"},
		{contractretirement.FaultMissing, "missing"},
		{contractretirement.FaultIdentityMismatch, "identity"},
		{contractretirement.FaultRetired, "retired"},
		{contractretirement.FaultOverflow, "overflow"},
		{contractretirement.FaultRevisionMismatch, "revision"},
		{contractretirement.FaultNotEligible, "not_eligible"},
	} {
		if contractretirement.FaultIs(err, candidate.kind) {
			return candidate.name
		}
	}
	return "untyped"
}

func verifyContractRetirementBoundary(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	authority *ContractRetirementDatabase, identity contractretirement.Identity,
	key contractretirement.Key, revision contractretirement.Revision,
) {
	t.Helper()
	value := keyValue(t, key)
	_, err := database.admin.Exec(ctx, `UPDATE platform.contract_retirements
SET zero_hit_since=clock_timestamp()-interval '30 days'+interval '1 second',
revision=revision+1,updated_at=clock_timestamp() WHERE contract_key=$1`, value)
	if err != nil {
		t.Fatalf("set pre-boundary: %s", err.Error())
	}
	revision, _ = contractretirement.NewRevision(revisionValue(t, revision) + 1)
	_, err = authority.MarkEligible(ctx, key, identity, revision)
	if !contractretirement.FaultIs(err, contractretirement.FaultNotEligible) {
		t.Fatalf("pre-boundary fault = %v", err)
	}
	_, err = database.admin.Exec(ctx, `UPDATE platform.contract_retirements
SET zero_hit_since=clock_timestamp()-interval '30 days',
revision=revision+1,updated_at=clock_timestamp() WHERE contract_key=$1`, value)
	if err != nil {
		t.Fatalf("set exact boundary: %s", err.Error())
	}
	revision, _ = contractretirement.NewRevision(revisionValue(t, revision) + 1)
	revision, err = authority.MarkEligible(ctx, key, identity, revision)
	if err != nil {
		t.Fatalf("mark eligible: %s", err.Error())
	}
	revision, err = authority.Finalize(ctx, key, identity, revision)
	if err != nil {
		t.Fatalf("finalize: %s", err.Error())
	}
	_, err = database.admin.Exec(ctx, `UPDATE platform.contract_retirements
SET revision=revision+1,updated_at=clock_timestamp() WHERE contract_key=$1`, value)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "55000" {
		t.Fatalf("terminal guard error = %v", err)
	}
	_, err = authority.Provision(ctx, key, identity, contractRetirementCount(t, 0))
	if !contractretirement.FaultIs(err, contractretirement.FaultRetired) {
		t.Fatalf("retired reintroduction fault = %v", err)
	}
}

func verifyEligibleHitReset(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	runtime *ContractRetirementDatabase, identity contractretirement.Identity,
) {
	t.Helper()
	key := contractRetirementKey(t, "f")
	value := keyValue(t, key)
	_, err := database.admin.Exec(ctx, `INSERT INTO platform.contract_retirements (
contract_key,owner_module,domain,state,remaining_scope_count,hit_count,
zero_hit_since,eligible_at,revision,updated_at)
VALUES ($1,'restquery','forms','eligible',0,0,clock_timestamp()-interval '31 days',
clock_timestamp(),1,clock_timestamp())`, value)
	if err != nil {
		t.Fatalf("seed eligible: %s", err.Error())
	}
	if _, err := runtime.RecordHits(
		ctx, contractRetirementBatch(t, identity, key),
	); err != nil {
		t.Fatalf("record eligible hit: %s", err.Error())
	}
	var state string
	var zero, eligible bool
	err = database.admin.QueryRow(ctx, `SELECT state,zero_hit_since IS NULL,
eligible_at IS NULL FROM platform.contract_retirements WHERE contract_key=$1`, value).Scan(
		&state, &zero, &eligible,
	)
	if err != nil || state != "active" || !zero || !eligible {
		t.Fatalf("eligible hit reset = %s/%t/%t err=%v", state, zero, eligible, err)
	}
}

func verifyContractRetirementOverflow(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	runtime *ContractRetirementDatabase, identity contractretirement.Identity,
) {
	t.Helper()
	key := contractRetirementKey(t, "9")
	_, err := database.admin.Exec(ctx, `INSERT INTO platform.contract_retirements (
contract_key,owner_module,domain,state,remaining_scope_count,hit_count,revision)
VALUES ($1,'restquery','forms','active',1,9223372036854775807,1)`, keyValue(t, key))
	if err != nil {
		t.Fatalf("seed overflow: %s", err.Error())
	}
	_, err = runtime.RecordHits(ctx, contractRetirementBatch(t, identity, key))
	if !contractretirement.FaultIs(err, contractretirement.FaultOverflow) {
		t.Fatalf("overflow fault = %v", err)
	}
}
