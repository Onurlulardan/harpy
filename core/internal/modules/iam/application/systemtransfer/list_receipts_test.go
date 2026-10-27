package systemtransfer

import (
	"context"
	"testing"
)

func TestListReceiptsDispatchesBoundedKeysetQuery(t *testing.T) {
	read := controlFixtureRead()
	store := &controlStoreFixture{readReceipts: func(got ControlRead,
		after int64, limit int) (ReceiptPage, error) {
		if got != read || after != 9 || limit != 20 {
			t.Fatal("receipt page input changed")
		}
		return ReceiptPage{NextAfter: 9}, nil
	}}
	result, err := ListReceipts(context.Background(), store, read, 9, 20)
	if err != nil || result.NextAfter != 9 {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
