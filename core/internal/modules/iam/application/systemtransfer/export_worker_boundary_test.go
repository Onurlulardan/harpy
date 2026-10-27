package systemtransfer

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestExportWorkerSealsOnlyAfterHistoricalSnapshotCloses(t *testing.T) {
	now := time.Date(2026, time.July, 18, 9, 0, 0, 123000000, time.UTC)
	attempt := fixtureExportAttempt(t, now)
	rows := fixtureExportRows()
	spool := &exportSpoolFixture{}
	snapshots := &exportSnapshotsFixture{snapshot: fixtureExportSnapshot(attempt, now, rows)}
	store := &exportStoreFixture{spool: spool}
	sealer := &exportSealerFixture{snapshots: snapshots, spool: spool,
		rows: rows, store: store, now: now}
	spools := &exportSpoolsFixture{spool: spool}
	worker, err := NewSystemExportWorker(exportClockFixture{now: now},
		exportLoaderFixture{attempt: attempt}, &exportLeaseFixture{},
		exportScopesFixture{}, exportAuthorizerFixture{},
		exportStoreBinderFixture{store: store}, snapshots,
		exportIdentityFixture{identity: fixtureExportReceiptID}, spools, sealer)
	if err != nil {
		t.Fatal(err)
	}
	result, err := worker.ExecuteExport(t.Context(), fixtureExportClaim())
	if err != nil || result.State != ExportExecutionReady || result.Code != "" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if sealer.called != 1 || !snapshots.closed || !rows.eof ||
		!spool.writerClosed || !spool.readerClosed || !spool.discarded {
		t.Fatalf("sealer=%d snapshot_closed=%v eof=%v spool=%+v",
			sealer.called, snapshots.closed, rows.eof, spool)
	}
	if len(store.starts) != 1 || store.completed != 1 ||
		store.completion.Progress.Rows != 1 ||
		store.completion.Progress.PlaintextBytes != int64(len(sealer.plaintext)) {
		t.Fatalf("completion=%+v", store.completion)
	}
	if !bytes.Contains(sealer.plaintext, []byte("person@example.com")) {
		t.Fatal("sealed plaintext did not contain the serialized snapshot")
	}
	if spools.identity.TenantID != attempt.TenantID ||
		spools.identity.ExportID != attempt.ExportID ||
		spools.identity.AttemptGeneration != fixtureExportClaim().LeaseGeneration ||
		spools.maximum != exportPlaintextMaximumBytes() {
		t.Fatalf("spool identity=%+v maximum=%d", spools.identity, spools.maximum)
	}
}

func TestExportWorkerPermanentlyRejectsStaleHistoricalCatalogPin(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ExportSnapshot)
	}{
		{name: "catalog version", mutate: func(snapshot *ExportSnapshot) {
			snapshot.CatalogVersionID = uuid.MustParse(
				"019be000-0000-7000-8000-000000000011")
		}},
		{name: "catalog digest", mutate: func(snapshot *ExportSnapshot) {
			snapshot.CatalogSHA256 = [32]byte{9}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now := time.Date(2026, time.July, 18, 9, 0, 0, 123000000, time.UTC)
			attempt := fixtureExportAttempt(t, now)
			rows := fixtureExportRows()
			spool := &exportSpoolFixture{}
			snapshot := fixtureExportSnapshot(attempt, now, rows)
			test.mutate(&snapshot)
			snapshots := &exportSnapshotsFixture{snapshot: snapshot}
			store := &exportStoreFixture{spool: spool}
			sealer := &exportSealerFixture{snapshots: snapshots, spool: spool,
				rows: rows, store: store, now: now}
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
			if executeErr != nil || result.State != ExportExecutionFailed ||
				result.Code != ExportCodeCatalogPinStale {
				t.Fatalf("result=%+v err=%v", result, executeErr)
			}
			if len(store.finished) != 1 ||
				store.finished[0].Code != "IAM_SYSTEM_EXPORT_CATALOG_PIN_STALE" ||
				sealer.called != 0 || !snapshots.closed || !spool.discarded {
				t.Fatalf("finished=%+v sealer=%d closed=%v discarded=%v",
					store.finished, sealer.called, snapshots.closed, spool.discarded)
			}
		})
	}
}

func TestNewSystemExportWorkerRejectsTypedNilPlaintextSpool(t *testing.T) {
	var spools *exportSpoolsFixture
	worker, err := NewSystemExportWorker(exportClockFixture{}, exportLoaderFixture{},
		&exportLeaseFixture{}, exportScopesFixture{}, exportAuthorizerFixture{},
		exportStoreBinderFixture{}, &exportSnapshotsFixture{},
		exportIdentityFixture{identity: fixtureExportReceiptID}, spools,
		&exportSealerFixture{})
	if err != ErrExportWorkerInvalid || worker != nil {
		t.Fatalf("worker=%v err=%v", worker, err)
	}
}
