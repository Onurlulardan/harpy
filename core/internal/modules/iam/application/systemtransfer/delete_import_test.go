package systemtransfer

import (
	"context"
	"testing"
	"time"
)

func TestDeleteImportDispatchesValidatedCommand(t *testing.T) {
	read, now := controlFixtureRead(), controlFixtureTime()
	store := &controlStoreFixture{cancelImport: func(got ControlRead,
		gotNow time.Time) (ImportRecord, error) {
		if got != read || gotNow != now {
			t.Fatal("cancel import input changed")
		}
		return ImportRecord{ImportID: read.TransferID}, nil
	}}
	result, err := DeleteImport(context.Background(), store, read, now)
	if err != nil || result.ImportID != read.TransferID {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
