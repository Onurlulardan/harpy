//go:build integration

package v1adoption

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

func TestWorkPG18FreshAndAdoptedCrashReplayIsolation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startWorkPG18(t, ctx)
	fixtures := prepareWorkSchemas(t, ctx, database.owner)
	seedFreshCanonical(t, ctx, database.owner, "work_fresh_a", "Fresh A", 11)
	seedFreshCanonical(t, ctx, database.owner, "work_fresh_b", "Fresh B", 12)
	assertSchemaBoard(t, ctx, database.owner, "work_fresh_a", "Fresh A", "Fresh B")
	assertSchemaBoard(t, ctx, database.owner, "work_fresh_b", "Fresh B", "Fresh A")

	runA := domain.ID(workUUID(31, 1))
	requestA := workAdoptionRequest(runA)
	runAdoptionTransaction(t, ctx, database.owner, "work_adopted_a",
		fixtures["work_adopted_a"], requestA, true)
	assertAdoptionRolledBack(t, ctx, database.owner, "work_adopted_a")
	evidenceA := runAdoptionTransaction(t, ctx, database.owner, "work_adopted_a",
		fixtures["work_adopted_a"], requestA, false)
	if !evidenceA.Accepted || !evidenceA.WriterFenced || evidenceA.SourceCount != 3 ||
		evidenceA.TargetCount != 8 {
		t.Fatalf("adopted A evidence = %#v", evidenceA)
	}

	requestB := workAdoptionRequest(domain.ID(workUUID(32, 1)))
	evidenceB := runAdoptionTransaction(t, ctx, database.owner, "work_adopted_b",
		fixtures["work_adopted_b"], requestB, false)
	if !evidenceB.Accepted || evidenceB.SourceCount != 3 || evidenceB.TargetCount != 8 {
		t.Fatalf("adopted B evidence = %#v", evidenceB)
	}
	assertSchemaBoard(t, ctx, database.owner, "work_adopted_a", "Adopted A", "Adopted B")
	assertSchemaBoard(t, ctx, database.owner, "work_adopted_b", "Adopted B", "Adopted A")
	assertAdoptedTaskMetadata(t, ctx, database.owner, "work_adopted_a",
		fixtures["work_adopted_a"])
	assertAdoptedTaskMetadata(t, ctx, database.owner, "work_adopted_b",
		fixtures["work_adopted_b"])
	assertJobReceiptCrashReplay(t, ctx, database.owner, "work_adopted_a")
}

func assertAdoptedTaskMetadata(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, fixture legacyFixture,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		var recurring bool
		var rule string
		var filter sql.NullString
		var end *time.Time
		err := transaction.QueryRowContext(ctx, `
SELECT is_recurring,recurrence_rule,recurrence_end,form_filter::text
FROM work_task_compatibility_metadata WHERE task_id=$1`, fixture.task).Scan(
			&recurring, &rule, &end, &filter)
		filterMatches := fixture.formFilter == nil && !filter.Valid || fixture.formFilter != nil &&
			filter.Valid && filter.String == `{"status": "open"}`
		if err != nil || !recurring || rule != "FREQ=DAILY;COUNT=3" || end == nil ||
			!filterMatches {
			t.Fatalf("adopted task metadata recurring=%t rule=%q end=%v filter=%q err=%v",
				recurring, rule, end, filter.String, err)
		}
	})
}

func runAdoptionTransaction(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, fixture legacyFixture, request adoption.ApplyRequest,
	rollback bool,
) adoption.Evidence {
	t.Helper()
	transaction, err := beginWorkSchema(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	service, scope, err := workAdoptionService(transaction, fixture)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	receipt, err := service.Copy(ctx, scope, request)
	if err != nil || receipt.State != "RECONCILING" || receipt.TargetCount != 8 {
		_ = transaction.Rollback()
		t.Fatalf("copy %s = %#v, %v", schema, receipt, err)
	}
	if rollback {
		if err = transaction.Rollback(); err != nil {
			t.Fatal(err)
		}
		return adoption.Evidence{}
	}
	evidence, err := service.Reconcile(ctx, scope, request)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("reconcile %s: %v", schema, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func workAdoptionRequest(runID domain.ID) adoption.ApplyRequest {
	return adoption.ApplyRequest{RunID: runID, Generation: 9,
		ManifestSHA256: sha256.Sum256([]byte("work-phase10-v1-adoption"))}
}
