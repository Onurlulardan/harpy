//go:build integration

package v1cutover

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
	"github.com/google/uuid"
)

func TestNotificationsPG18CutoverEvidenceFenceCASAndTenantIsolation(t *testing.T) {
	ctx := context.Background()
	connection := startCutoverPG18(t, ctx)
	prepareCutoverDatabase(t, ctx, connection)
	tenant := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	runID := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	transaction, err := connection.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	if _, err = transaction.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		"cutover_a, pg_catalog"); err != nil {
		t.Fatal(err)
	}
	store := cutoverTestStore(transaction, tenant)
	authority, err := store.CreateAuthority(ctx, app.OriginV1Adopted)
	if err != nil || authority.Writer != "v1" {
		t.Fatalf("bootstrap=%+v err=%v", authority, err)
	}
	seedVerifiedRun(t, ctx, transaction, runID)
	stale := authority
	stale.RowVersion--
	if _, err = store.PauseV1(ctx, stale); err == nil {
		t.Fatal("stale authority quiesced")
	}
	authority, err = store.PauseV1(ctx, authority)
	if err != nil || authority.State != "quiescing" {
		t.Fatalf("quiesce=%+v err=%v", authority, err)
	}
	nonzero := cutoverInventory(runID, 1, "nonzero")
	evidence, err := store.RecordEvidence(ctx, nonzero)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.DisableV1(ctx, authority, evidence); err == nil {
		t.Fatal("non-zero inventory fenced")
	}
	zero := cutoverInventory(runID, 0, "zero")
	evidence, err = store.RecordEvidence(ctx, zero)
	if err != nil {
		t.Fatal(err)
	}
	authority, err = store.DisableV1(ctx, authority, evidence)
	if err != nil || authority.State != "fenced" {
		t.Fatalf("fence=%+v err=%v", authority, err)
	}
	authority, err = store.ActivateV2(ctx, authority)
	if err != nil || authority.Writer != "v2" || authority.Generation != 2 {
		t.Fatalf("activate=%+v err=%v", authority, err)
	}
	if err = transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var otherWriter string
	if err = connection.QueryRow(ctx, `SELECT writer FROM cutover_b.notifications_writer_authority`).Scan(&otherWriter); err != nil || otherWriter != "unassigned" {
		t.Fatalf("other tenant writer=%s err=%v", otherWriter, err)
	}
}

func cutoverInventory(runID uuid.UUID, queue int64, key string) app.Inventory {
	return app.Inventory{RunID: runID, SourceQueueDepth: queue,
		SourceEvidenceDigest: sha256.Sum256([]byte(key)), ObservedAt: time.Now().UTC().Add(-time.Second)}
}
