//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
)

func TestBillingFormsGroupRehearsalAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			database := startCutoverIntegrationDatabaseForOrigin(t, ctx, origin)
			value := groupRehearsalFixture()
			seedBillingFormsRehearsalGroup(t, ctx, database.admin, value, true)
			leaseGroupRehearsalAdopter(t, ctx, database.admin)
			adopter := connectCutoverRole(t, ctx, database.dsn, "harpy_adopter")
			defer adopter.Close(context.Background())
			recorder, err := NewGroupRehearsalRecorder(adopter)
			if err != nil {
				t.Fatal(err)
			}
			accepted, err := grouprehearsal.Record(ctx, recorder, value)
			if err != nil {
				t.Fatal(err)
			}
			digest, err := accepted.SHA256()
			if err != nil || digest == [sha256.Size]byte{} {
				t.Fatalf("billing Forms acceptance digest=%x err=%v", digest, err)
			}
			if _, err := grouprehearsal.Record(ctx, recorder, value); err != nil {
				t.Fatalf("exact rehearsal replay: %v", err)
			}
			assertPersistedBillingFormsRehearsal(t, ctx, database, value, digest)
			_, err = adopter.Exec(ctx, `UPDATE platform.cutover_billing_forms_rehearsals
SET usage_flushed=false WHERE receipt_id=$1`, value.ReceiptID)
			requireGroupRehearsalSQLState(t, err, "42501")
			assertBillingFormsRehearsalNegatives(t, ctx, database, recorder, value, digest)
		})
	}
}

func assertPersistedBillingFormsRehearsal(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	value grouprehearsal.PersistedRehearsal, digest [sha256.Size]byte,
) {
	t.Helper()
	var groupID, persistedDigest string
	var generation, billingCount, formsCount int64
	var quiesced, flushed, dualWrite bool
	err := database.admin.QueryRow(ctx, `SELECT group_id,group_generation,
billing_source_count,forms_source_count,quiesce_drained,usage_flushed,
production_dual_write,encode(group_sha256,'hex')
FROM platform.cutover_billing_forms_rehearsals WHERE receipt_id=$1`, value.ReceiptID).Scan(
		&groupID, &generation, &billingCount, &formsCount,
		&quiesced, &flushed, &dualWrite, &persistedDigest)
	if err != nil || groupID != grouprehearsal.CanonicalGroupID ||
		generation != value.Evidence.Generation || billingCount != 2 || formsCount != 4 ||
		!quiesced || !flushed || dualWrite || persistedDigest != fmtDigest(digest) {
		t.Fatalf("persisted billing Forms receipt invalid: group=%s generation=%d digest=%s err=%v",
			groupID, generation, persistedDigest, err)
	}
}

func assertBillingFormsRehearsalNegatives(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	recorder grouprehearsal.Recorder, value grouprehearsal.PersistedRehearsal,
	digest [sha256.Size]byte,
) {
	t.Helper()
	tampered := digest
	tampered[0] ^= 0xff
	_, err := recorder.Record(ctx, value, tampered)
	requireGroupRehearsalSQLState(t, err, "P57B3")
	stale := value
	stale.Evidence.Generation++
	_, err = grouprehearsal.Record(ctx, recorder, stale)
	requireGroupRehearsalSQLState(t, err, "P57B1")
	missing := groupRehearsalWithIdentity(
		"20000000-0000-4000-8000-000000000009",
		"20000000-0000-4000-8000-000000000010")
	_, err = grouprehearsal.Record(ctx, recorder, missing)
	requireGroupRehearsalSQLState(t, err, "P0002")
	partial := groupRehearsalWithIdentity(
		"20000000-0000-4000-8000-000000000011",
		"20000000-0000-4000-8000-000000000012")
	seedBillingFormsRehearsalGroup(t, ctx, database.admin, partial, false)
	_, err = grouprehearsal.Record(ctx, recorder, partial)
	requireGroupRehearsalSQLState(t, err, "P57B2")
	_, err = database.admin.Exec(ctx, `UPDATE platform.cutover_billing_forms_rehearsals
SET usage_flushed=false WHERE receipt_id=$1`, value.ReceiptID)
	requireGroupRehearsalSQLState(t, err, "55000")
}

func fmtDigest(value [sha256.Size]byte) string {
	const hexadecimal = "0123456789abcdef"
	result := make([]byte, sha256.Size*2)
	for index, octet := range value {
		result[index*2], result[index*2+1] = hexadecimal[octet>>4], hexadecimal[octet&15]
	}
	return string(result)
}
