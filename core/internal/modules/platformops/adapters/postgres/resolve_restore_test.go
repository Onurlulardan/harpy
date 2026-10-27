package postgres

import (
	"context"
	"testing"
	"time"
)

func TestResolveRestore(t *testing.T) {
	now := time.Date(2026, 7, 23, 1, 0, 0, 0, time.UTC)
	database := &operationDatabase{row: func(values ...any) error {
		for index, value := range []string{"operation-id", "set-id", "environment-id", "empty", "verifying", "plan"} {
			*(values[index].(*string)) = value
		}
		receipt, databaseID := "receipt", "database-id"
		*(values[6].(**string)) = &receipt
		*(values[7].(**string)) = &databaseID
		*(values[8].(*time.Time)) = now
		*(values[9].(*time.Time)) = now
		*(values[10].(**time.Time)) = &now
		return nil
	}}
	record, err := operationStore(database).ResolveRestore(context.Background(), "operation-id")
	if err != nil || record.OperationID != "operation-id" || record.ReceiptSHA256 != "receipt" {
		t.Fatalf("get restore = %#v %v", record, err)
	}
}
