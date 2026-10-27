package systemtransfer

import (
	"context"
	"testing"
	"time"
)

func TestCancelExportDispatchesValidatedCommand(t *testing.T) {
	read, now := controlFixtureRead(), controlFixtureTime()
	store := &controlStoreFixture{cancelExport: func(got ControlRead,
		gotNow time.Time) (ExportRecord, error) {
		if got != read || gotNow != now {
			t.Fatal("cancel export input changed")
		}
		return ExportRecord{ExportID: read.TransferID}, nil
	}}
	result, err := CancelExport(context.Background(), store, read, now)
	if err != nil || result.ExportID != read.TransferID {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
