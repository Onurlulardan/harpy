package systemtransfer

import (
	"context"
	"testing"
	"time"
)

func TestDownloadExportArtifactDispatchesFreshQuery(t *testing.T) {
	read, now := controlFixtureRead(), controlFixtureTime()
	store := &controlStoreFixture{readExportArtifact: func(got ControlRead,
		gotNow time.Time) (ExportArtifact, error) {
		if got != read || gotNow != now {
			t.Fatal("artifact query input changed")
		}
		return ExportArtifact{ExportRecord: ExportRecord{ExportID: read.TransferID}}, nil
	}}
	result, err := DownloadExportArtifact(context.Background(), store, read, now)
	if err != nil || result.ExportID != read.TransferID {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
