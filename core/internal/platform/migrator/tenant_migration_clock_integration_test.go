//go:build integration

package migrator

import (
	"context"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestTenantCheckpointPostgreSQL184RejectsCallerClockAndResetsRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	selected := runner.tenantSelected[0]
	tenantID := "ffffffff-0000-4000-8000-000000000006"
	_, err := fixture.database.ExecContext(ctx, `INSERT INTO platform.tenant_migration_checkpoints
(tenant_id,owner,logical_version,database_origin,selected_path,selected_sha256,
run_id,state,attempt,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,'pending',1,
TIMESTAMPTZ '2000-01-01 00:00:00+00')`, tenantID, selected.Owner, selected.Version,
		selected.Origin, selected.Path, selected.SHA256, tenantRunOne)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fixture.database.ExecContext(ctx, `UPDATE platform.tenant_migration_checkpoints
SET state='running',started_at=TIMESTAMPTZ '2000-01-01 00:00:00+00',
updated_at=TIMESTAMPTZ '2000-01-01 00:00:00+00'
WHERE tenant_id=$1 AND owner=$2 AND logical_version=$3`, tenantID, selected.Owner, selected.Version)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fixture.database.ExecContext(ctx, `INSERT INTO platform.tenant_migration_checkpoint_attempts
(tenant_id,owner,logical_version,attempt,database_origin,selected_path,
selected_sha256,run_id,state,started_at,finished_at,error_code,recorded_at)
VALUES($1,$2,$3,1,$4,$5,$6,$7,'failed',TIMESTAMPTZ '1999-01-01 00:00:00+00',
TIMESTAMPTZ '1999-01-01 00:00:00+00','migration_execution_failed',
TIMESTAMPTZ '2099-01-01 00:00:00+00')`, tenantID, selected.Owner, selected.Version,
		selected.Origin, selected.Path, selected.SHA256, tenantRunOne)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fixture.database.ExecContext(ctx, `UPDATE platform.tenant_migration_checkpoints
SET state='failed',finished_at=TIMESTAMPTZ '2000-01-01 00:00:00+00',
error_code='migration_execution_failed',updated_at=TIMESTAMPTZ '2000-01-01 00:00:00+00'
WHERE tenant_id=$1 AND owner=$2 AND logical_version=$3`, tenantID, selected.Owner, selected.Version)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fixture.database.ExecContext(ctx, `UPDATE platform.tenant_migration_checkpoints
SET state='running',run_id=$4,attempt=2,
started_at=TIMESTAMPTZ '2000-01-01 00:00:00+00',finished_at=TIMESTAMPTZ '2000-01-01 00:00:00+00',
error_code='migration_execution_failed',updated_at=TIMESTAMPTZ '2000-01-01 00:00:00+00'
WHERE tenant_id=$1 AND owner=$2 AND logical_version=$3`, tenantID, selected.Owner,
		selected.Version, tenantRunTwo)
	if err != nil {
		t.Fatal(err)
	}
	var current bool
	var finished, errorCode *string
	if err := fixture.database.QueryRowContext(ctx, `SELECT
started_at > TIMESTAMPTZ '2026-01-01 00:00:00+00',finished_at::text,error_code
FROM platform.tenant_migration_checkpoints WHERE tenant_id=$1`, tenantID).Scan(
		&current, &finished, &errorCode,
	); err != nil || !current || finished != nil || errorCode != nil {
		t.Fatalf("retry clock/reset = current %v finished %v error %v err=%v",
			current, finished, errorCode, err)
	}
	var callerTimeSurvived bool
	if err := fixture.database.QueryRowContext(ctx, `SELECT
started_at=TIMESTAMPTZ '1999-01-01 00:00:00+00'
OR finished_at=TIMESTAMPTZ '1999-01-01 00:00:00+00'
OR recorded_at=TIMESTAMPTZ '2099-01-01 00:00:00+00'
FROM platform.tenant_migration_checkpoint_attempts WHERE tenant_id=$1`, tenantID).Scan(
		&callerTimeSurvived,
	); err != nil || callerTimeSurvived {
		t.Fatalf("caller clock survived = %v err=%v", callerTimeSurvived, err)
	}
}
