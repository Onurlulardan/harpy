package systemtransfer

import (
	"context"
	"testing"
)

func TestGetExportDispatchesValidatedQuery(t *testing.T) {
	read := controlFixtureRead()
	store := &controlStoreFixture{readExport: func(got ControlRead) (ExportRecord, error) {
		if got != read {
			t.Fatal("read export input changed")
		}
		return ExportRecord{ExportID: read.TransferID}, nil
	}}
	result, err := GetExport(context.Background(), store, read)
	if err != nil || result.ExportID != read.TransferID {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
