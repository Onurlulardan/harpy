package systemtransfer

import (
	"context"
	"testing"
)

func TestResolveImportDispatchesValidatedQuery(t *testing.T) {
	read := controlFixtureRead()
	store := &controlStoreFixture{readImport: func(got ControlRead) (ImportRecord, error) {
		if got != read {
			t.Fatal("read import input changed")
		}
		return ImportRecord{ImportID: read.TransferID}, nil
	}}
	result, err := ResolveImport(context.Background(), store, read)
	if err != nil || result.ImportID != read.TransferID {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
