package analyticsdashboardexport

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type testRow func(...any) error

func (row testRow) Scan(destinations ...any) error { return row(destinations...) }

type testDatabase struct{ row platformpostgres.Row }

func (database testDatabase) QueryRow(
	context.Context, string, ...any,
) platformpostgres.Row {
	return database.row
}

func exportTestID(suffix string) uuid.UUID {
	return uuid.MustParse("01900000-0000-7000-8000-" + suffix)
}

func TestCandidateAndUnknownCandidateDecodeTypedEvidence(t *testing.T) {
	tenantID, exportID := exportTestID("000000000001"), exportTestID("000000000002")
	jobID, attemptID := exportTestID("000000000003"), exportTestID("000000000004")
	leaseID, effectID := exportTestID("000000000005"), exportTestID("000000000006")
	store := &Store{database: testDatabase{row: testRow(func(out ...any) error {
		*out[0].(*uuid.UUID), *out[1].(*string) = tenantID, "tenant_a"
		*out[2].(*uuid.UUID), *out[3].(*uuid.UUID) = exportID, jobID
		*out[4].(*int64) = 1
		return nil
	})}}
	candidate, found, err := store.NextAnalyticsDashboardExport(context.Background())
	if err != nil || !found || candidate.ExportID != exportID {
		t.Fatalf("candidate decode failed: found=%v candidate=%+v err=%v", found, candidate, err)
	}
	digest := make([]byte, 32)
	digest[0] = 7
	store.database = testDatabase{row: testRow(func(out ...any) error {
		*out[0].(*uuid.UUID), *out[1].(*string) = tenantID, "tenant_a"
		*out[2].(*uuid.UUID), *out[3].(*uuid.UUID) = exportID, jobID
		*out[4].(*uuid.UUID), *out[5].(*int64) = attemptID, 1
		*out[6].(*uuid.UUID), *out[7].(*uuid.UUID) = leaseID, effectID
		*out[8].(*[]byte) = digest
		*out[9].(*time.Time) = time.Unix(1, 0).UTC()
		return nil
	})}
	unknown, found, err := store.NextAnalyticsDashboardExportUnknown(context.Background())
	if err != nil || !found || unknown.AttemptID != attemptID ||
		unknown.EffectRequestSHA256[0] != 7 {
		t.Fatalf("unknown decode failed: found=%v candidate=%+v err=%v", found, unknown, err)
	}
}

func TestSourceNoRowAndSpoolTenantNullAreNotErrors(t *testing.T) {
	store := &Store{database: testDatabase{row: testRow(func(...any) error {
		return pgx.ErrNoRows
	})}}
	if _, found, err := store.NextAnalyticsDashboardExport(context.Background()); err != nil || found {
		t.Fatalf("no candidate mismatch: found=%v err=%v", found, err)
	}
	store.database = testDatabase{row: testRow(func(out ...any) error {
		*out[0].(*pgtype.UUID) = pgtype.UUID{}
		return nil
	})}
	if _, found, err := store.AnalyticsDashboardExportSpoolTenant(
		context.Background(), exportTestID("000000000002"),
	); err != nil || found {
		t.Fatalf("null spool tenant mismatch: found=%v err=%v", found, err)
	}
}

func TestUnknownCandidateRejectsInvalidDigest(t *testing.T) {
	store := &Store{database: testDatabase{row: testRow(func(out ...any) error {
		*out[0].(*uuid.UUID), *out[1].(*string) = exportTestID("000000000001"), "tenant_a"
		for _, index := range []int{2, 3, 4, 6, 7} {
			*out[index].(*uuid.UUID) = exportTestID("000000000002")
		}
		*out[5].(*int64), *out[8].(*[]byte) = 1, []byte{1}
		*out[9].(*time.Time) = time.Unix(1, 0).UTC()
		return nil
	})}}
	_, _, err := store.NextAnalyticsDashboardExportUnknown(context.Background())
	if !errors.Is(err, platformasync.ErrAnalyticsDashboardExportEvidenceInvalid) {
		t.Fatalf("invalid digest error = %v", err)
	}
}

func TestExpiredCandidateDecodesPostEffectEvidence(t *testing.T) {
	tenantID, exportID := exportTestID("000000000001"), exportTestID("000000000002")
	jobID, attemptID := exportTestID("000000000003"), exportTestID("000000000004")
	leaseID, effectID := exportTestID("000000000005"), exportTestID("000000000006")
	digest := make([]byte, 32)
	digest[0] = 9
	store := &Store{database: testDatabase{row: testRow(func(out ...any) error {
		*out[0].(*uuid.UUID), *out[1].(*string) = tenantID, "tenant_a"
		*out[2].(*uuid.UUID), *out[3].(*uuid.UUID) = exportID, jobID
		*out[4].(*uuid.UUID), *out[5].(*int64) = attemptID, 1
		*out[6].(*uuid.UUID), *out[7].(*bool) = leaseID, true
		*out[8].(*pgtype.UUID) = pgtype.UUID{Bytes: effectID, Valid: true}
		*out[9].(*[]byte), *out[10].(*bool) = digest, true
		*out[11].(*time.Time) = time.Unix(2, 0).UTC()
		return nil
	})}}
	candidate, found, err := store.NextAnalyticsDashboardExportExpired(
		context.Background(),
	)
	if err != nil || !found || candidate.EffectID != effectID ||
		!candidate.CancellationPending || candidate.EffectRequestSHA256[0] != 9 {
		t.Fatalf("expired decode failed: found=%v candidate=%+v err=%v", found, candidate, err)
	}
}

func TestMaintenanceUnknownCandidateDecodesFrozenAuthorityEvidence(t *testing.T) {
	tenantID, exportID := exportTestID("000000000001"), exportTestID("000000000002")
	jobID, attemptID := exportTestID("000000000003"), exportTestID("000000000004")
	leaseID := exportTestID("000000000005")
	effectID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("maintenance-effect"))
	digest := make([]byte, 32)
	digest[0] = 11
	row := func(leaseGeneration int64) testRow {
		return func(out ...any) error {
			*out[0].(*string), *out[1].(*uuid.UUID) = platformasync.AnalyticsDashboardExportDeleteJobKind, tenantID
			*out[2].(*string), *out[3].(*uuid.UUID) = "tenant_a", exportID
			*out[4].(*string), *out[5].(*int64), *out[6].(*int64) = "analytics", 4, 5
			*out[7].(*uuid.UUID), *out[8].(*uuid.UUID) = jobID, attemptID
			*out[9].(*int64), *out[10].(*int64) = 1, leaseGeneration
			*out[11].(*uuid.UUID), *out[12].(*uuid.UUID) = leaseID, effectID
			*out[13].(*[]byte), *out[14].(*time.Time) = digest, time.Unix(3, 0).UTC()
			return nil
		}
	}
	store := &Store{database: testDatabase{row: row(1)}}
	candidate, found, err := store.NextAnalyticsDashboardExportMaintenanceUnknown(
		context.Background(),
	)
	if err != nil || !found || candidate.AuthorityGroupGeneration != 4 ||
		candidate.CutoverGeneration != 5 || candidate.EffectID != effectID ||
		candidate.EffectRequestSHA256[0] != 11 {
		t.Fatalf("maintenance candidate decode failed: found=%v candidate=%+v err=%v",
			found, candidate, err)
	}
	drifted := &Store{database: testDatabase{row: row(2)}}
	_, found, err = drifted.NextAnalyticsDashboardExportMaintenanceUnknown(
		context.Background(),
	)
	if found || !errors.Is(err, platformasync.ErrAnalyticsDashboardExportEvidenceInvalid) {
		t.Fatalf("maintenance lease drift found=%v err=%v", found, err)
	}
}

func TestSQLDocumentsCallOnlyTypedPlatformRoutines(t *testing.T) {
	for name, document := range map[string]string{
		"candidate": candidateSQL, "unknown": unknownCandidateSQL,
		"maintenance_unknown": maintenanceUnknownCandidateSQL,
		"expired":             expiredCandidateSQL, "spool": spoolTenantSQL,
	} {
		if !strings.Contains(document, "platform_async_analytics_dashboard_export_") ||
			strings.Contains(document, "platform.jobs") || strings.Contains(document, "job_attempts") {
			t.Fatalf("%s SQL bypasses typed platform routine: %s", name, document)
		}
	}
}
