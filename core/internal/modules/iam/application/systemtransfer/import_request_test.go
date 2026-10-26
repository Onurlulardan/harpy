package systemtransfer

import (
	"context"
	"testing"
)

func TestImportRequestDispatchesValidatedWrite(t *testing.T) {
	write := controlFixtureImportWrite()
	store := &controlStoreFixture{submitImport: func(got ImportWrite) (ImportRecord, error) {
		if got.ImportID != write.ImportID {
			t.Fatal("submit import input changed")
		}
		return ImportRecord{ImportID: got.ImportID}, nil
	}}
	result, err := ImportRequest(context.Background(), store, write)
	if err != nil || result.ImportID != write.ImportID {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
