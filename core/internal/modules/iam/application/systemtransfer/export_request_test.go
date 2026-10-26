package systemtransfer

import (
	"context"
	"testing"
)

func TestExportRequestDispatchesValidatedWrite(t *testing.T) {
	write := controlFixtureExportWrite()
	store := &controlStoreFixture{submitExport: func(got ExportWrite) (ExportRecord, error) {
		if got.ExportID != write.ExportID {
			t.Fatal("submit export input changed")
		}
		return ExportRecord{ExportID: got.ExportID}, nil
	}}
	result, err := ExportRequest(context.Background(), store, write)
	if err != nil || result.ExportID != write.ExportID {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
