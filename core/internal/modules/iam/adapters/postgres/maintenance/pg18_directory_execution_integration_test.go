//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func assertMaintenancePG18DirectoryExecution(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	seedMaintenancePG18Directory(t, ctx, database)
	jobs := []struct {
		tenant maintenancePG18Tenant
		repair string
		check  string
	}{
		{maintenancePG18Tenants[0], iamMaintenanceJobID, iamMaintenanceChecksumJobA},
		{maintenancePG18Tenants[1], iamMaintenanceRepairJobB, iamMaintenanceChecksumJobB},
	}
	for _, fixture := range jobs {
		bootstrapMaintenancePG18Schedule(t, ctx, database, fixture.tenant,
			maintenanceapp.KindDirectoryRepair, fixture.repair)
		bootstrapMaintenancePG18Schedule(t, ctx, database, fixture.tenant,
			maintenanceapp.KindDirectoryChecksum, fixture.check)
	}
	claimA := recurringMaintenancePG18Claim(t, jobs[0].tenant.id,
		jobs[0].repair, iamMaintenanceAttemptA, maintenanceapp.KindDirectoryRepair)
	claimB := recurringMaintenancePG18Claim(t, jobs[1].tenant.id,
		jobs[1].repair, iamMaintenanceAttemptB, maintenanceapp.KindDirectoryRepair)
	if _, err := executeMaintenancePG18Command(t, ctx, database,
		jobs[0].tenant.schema, claimB,
		uuid.MustParse(iamMaintenanceNextRepairB), true); err == nil {
		t.Fatal("foreign tenant directory repair claim passed")
	}
	bad := claimA
	bad.PayloadSHA256 = fmt.Sprintf("%064d", 0)
	if _, err := executeMaintenancePG18Command(t, ctx, database,
		jobs[0].tenant.schema, bad,
		uuid.MustParse(iamMaintenanceNextRepairA), true); err == nil {
		t.Fatal("invalid directory repair evidence passed")
	}
	rolled, err := executeMaintenancePG18Command(t, ctx, database,
		jobs[0].tenant.schema, claimA,
		uuid.MustParse(iamMaintenanceNextRepairA), false)
	if err != nil || rolled.Successor == nil {
		t.Fatalf("rollback directory repair=%#v/%v", rolled, err)
	}
	assertMaintenancePG18RepairRollback(t, ctx, database, jobs[0].tenant, claimA.JobID)
	assertMaintenancePG18RepairCycle(t, ctx, database, jobs[0].tenant,
		claimA, iamMaintenanceNextRepairA,
		"019c6a7e-7c00-7000-8000-000000000210")
	assertMaintenancePG18RepairCycle(t, ctx, database, jobs[1].tenant,
		claimB, iamMaintenanceNextRepairB,
		"019c6a7e-7c00-7000-8000-000000000220")
	for index, fixture := range jobs {
		assertMaintenancePG18ProjectionParity(t, ctx, database, fixture.tenant)
		next := []string{iamMaintenanceNextChecksumA, iamMaintenanceNextChecksumB}[index]
		attempt := []string{iamMaintenanceAttemptCheckA, iamMaintenanceAttemptCheckB}[index]
		claim := recurringMaintenancePG18Claim(t, fixture.tenant.id,
			fixture.check, attempt, maintenanceapp.KindDirectoryChecksum)
		outcome, err := executeMaintenancePG18Command(t, ctx, database,
			fixture.tenant.schema, claim, uuid.MustParse(next), true)
		if err != nil || outcome.Successor == nil {
			t.Fatalf("%s checksum=%#v/%v", fixture.tenant.schema, outcome, err)
		}
		assertMaintenancePG18ChecksumParity(t, ctx, database, fixture.tenant)
	}
}

func assertMaintenancePG18RepairCycle(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant, claim maintenanceapp.Claim,
	firstNext, resumeNext string,
) {
	t.Helper()
	outcome, err := executeMaintenancePG18Command(t, ctx, database, tenant.schema,
		claim, uuid.MustParse(firstNext), true)
	if err != nil || outcome.Successor == nil ||
		outcome.Successor.JobID != uuid.MustParse(firstNext) {
		t.Fatalf("%s repair=%#v/%v", tenant.schema, outcome, err)
	}
	replay, err := executeMaintenancePG18Command(t, ctx, database, tenant.schema,
		claim, uuid.MustParse(resumeNext), true)
	if err != nil || replay.Successor == nil ||
		replay.Successor.JobID != outcome.Successor.JobID {
		t.Fatalf("%s repair replay=%#v/%v", tenant.schema, replay, err)
	}
	resume := nextMaintenancePG18Claim(t, tenant.id,
		maintenanceapp.KindDirectoryRepair, outcome.Successor.JobID,
		iamMaintenanceAttemptResume)
	continued, err := executeMaintenancePG18Command(t, ctx, database,
		tenant.schema, resume, uuid.MustParse(resumeNext), true)
	if err != nil || continued.Successor == nil {
		t.Fatalf("%s repair resume=%#v/%v", tenant.schema, continued, err)
	}
}

func assertMaintenancePG18RepairRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant, jobID uuid.UUID,
) {
	t.Helper()
	var projection, receipts int
	err := database.QueryRowContext(ctx, `SELECT
 (SELECT pg_catalog.count(*) FROM platform.iam_login_directory
  WHERE tenant_id=$1 AND source_identifier_id=$2),
 (SELECT pg_catalog.count(*) FROM `+tenant.schema+`.iam_maintenance_receipts
  WHERE job_id=$3)`, tenant.id, iamMaintenanceSourceID, jobID).Scan(
		&projection, &receipts)
	if err != nil || projection != 0 || receipts != 0 {
		t.Fatalf("%s rollback=%d/%d: %v", tenant.schema, projection, receipts, err)
	}
}

func assertMaintenancePG18ProjectionParity(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant,
) {
	t.Helper()
	var sourceRevision, projectedRevision int64
	var sourceSHA, projectedSHA []byte
	err := database.QueryRowContext(ctx, `SELECT source.revision,
 projected.source_revision,source.normalized_sha256,projected.source_sha256
FROM `+tenant.schema+`.iam_user_identifiers AS source
JOIN platform.iam_login_directory AS projected
 ON projected.tenant_id=$1 AND projected.source_identifier_id=source.id
WHERE source.id=$2 AND projected.lookup_key_id=$3`,
		tenant.id, iamMaintenanceSourceID, iamMaintenanceKeyID).Scan(
		&sourceRevision, &projectedRevision, &sourceSHA, &projectedSHA)
	if err != nil || sourceRevision != projectedRevision ||
		!bytes.Equal(sourceSHA, projectedSHA) {
		t.Fatalf("%s projection parity=%d/%d/%v",
			tenant.schema, sourceRevision, projectedRevision, err)
	}
}

func assertMaintenancePG18ChecksumParity(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant,
) {
	t.Helper()
	var scanned, drift, projected int64
	err := database.QueryRowContext(ctx, `SELECT checkpoint.scanned_count,
 checkpoint.affected_count,evidence.projected_row_count
FROM `+tenant.schema+`.iam_maintenance_checkpoints AS checkpoint
JOIN `+tenant.schema+`.iam_directory_key_evidence AS evidence
 ON evidence.lookup_key_id=$1
WHERE checkpoint.kind='iam.directory.checksum'`, iamMaintenanceKeyID).Scan(
		&scanned, &drift, &projected)
	if err != nil || scanned != 1 || drift != 0 || projected != 1 {
		t.Fatalf("%s checksum parity=%d/%d/%d: %v",
			tenant.schema, scanned, drift, projected, err)
	}
}
