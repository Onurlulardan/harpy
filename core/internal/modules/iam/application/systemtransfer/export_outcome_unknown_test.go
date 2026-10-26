package systemtransfer

import (
	"errors"
	"testing"
	"time"
)

func TestExportWorkerPersistsProviderAndReadyOutcomeUnknown(t *testing.T) {
	tests := []struct {
		name          string
		afterBegin    error
		completeError error
		artifactKnown bool
	}{
		{name: "provider response lost",
			afterBegin: errors.New("provider response was lost")},
		{name: "ready commit response lost",
			completeError: errors.New("ready commit response was lost"),
			artifactKnown: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now := time.Date(2026, time.July, 18, 9, 0, 0, 123000000, time.UTC)
			attempt := fixtureExportAttempt(t, now)
			rows := fixtureExportRows()
			spool := &exportSpoolFixture{}
			snapshots := &exportSnapshotsFixture{
				snapshot: fixtureExportSnapshot(attempt, now, rows)}
			store := &exportStoreFixture{spool: spool,
				completeErr: test.completeError}
			sealer := &exportSealerFixture{snapshots: snapshots, spool: spool,
				rows: rows, store: store, now: now, afterBegin: test.afterBegin}
			worker, err := NewSystemExportWorker(exportClockFixture{now: now},
				exportLoaderFixture{attempt: attempt}, &exportLeaseFixture{},
				exportScopesFixture{}, exportAuthorizerFixture{},
				exportStoreBinderFixture{store: store}, snapshots,
				exportIdentityFixture{identity: fixtureExportReceiptID},
				&exportSpoolsFixture{spool: spool}, sealer)
			if err != nil {
				t.Fatal(err)
			}
			result, executeErr := worker.ExecuteExport(t.Context(), fixtureExportClaim())
			if executeErr != nil || result.State != ExportExecutionOutcomeUnknown ||
				result.Code != "JOB_OUTCOME_UNKNOWN" {
				t.Fatalf("result=%+v err=%v", result, executeErr)
			}
			if len(store.starts) != 1 || len(store.unknown) != 1 ||
				len(store.finished) != 0 || store.completed != 0 {
				t.Fatalf("starts=%d unknown=%d finished=%d completed=%d",
					len(store.starts), len(store.unknown),
					len(store.finished), store.completed)
			}
			start := store.starts[0]
			if start.EffectID != fixtureAttemptID ||
				start.EffectReceiptID != fixtureExportReceiptID ||
				start.EffectID == start.EffectReceiptID ||
				store.unknown[0].ArtifactKnown != test.artifactKnown ||
				store.unknown[0].Validate() != nil || !spool.discarded {
				t.Fatalf("start=%+v unknown=%+v discarded=%v",
					start, store.unknown[0], spool.discarded)
			}
		})
	}
}

func TestExportWorkerReturnsStoredOutcomeUnknownWithoutRetry(t *testing.T) {
	now := time.Date(2026, time.July, 18, 9, 0, 0, 123000000, time.UTC)
	attempt := fixtureExportAttempt(t, now)
	attempt.State = ExportOutcomeUnknown
	store := &exportStoreFixture{}
	sealer := &exportSealerFixture{}
	worker, err := NewSystemExportWorker(exportClockFixture{now: now},
		exportLoaderFixture{attempt: attempt}, &exportLeaseFixture{},
		exportScopesFixture{}, exportAuthorizerFixture{},
		exportStoreBinderFixture{store: store}, &exportSnapshotsFixture{},
		exportIdentityFixture{identity: fixtureExportReceiptID},
		&exportSpoolsFixture{}, sealer)
	if err != nil {
		t.Fatal(err)
	}
	result, err := worker.ExecuteExport(t.Context(), fixtureExportClaim())
	if err != nil || result.State != ExportExecutionOutcomeUnknown ||
		result.Code != ExportOutcomeUnknownCode || sealer.called != 0 ||
		len(store.finished) != 0 {
		t.Fatalf("result=%+v err=%v sealer=%d finished=%d",
			result, err, sealer.called, len(store.finished))
	}
}
