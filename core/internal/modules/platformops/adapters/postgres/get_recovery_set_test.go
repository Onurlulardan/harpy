package postgres

import (
	"context"
	"testing"
	"time"
)

func TestGetRecoverySet(t *testing.T) {
	now := time.Date(2026, 7, 23, 1, 0, 0, 0, time.UTC)
	database := &operationDatabase{row: func(values ...any) error {
		*(values[0].(*string)) = "set-id"
		*(values[1].(*string)) = "operation-id"
		*(values[2].(*string)) = "environment-id"
		*(values[3].(*int64)) = 4
		*(values[4].(*string)) = "protected"
		digest := "manifest"
		*(values[5].(**string)) = &digest
		*(values[6].(*[]byte)) = []byte("manifest")
		*(values[7].(*bool)) = false
		databaseID := "database-id"
		*(values[8].(**string)) = &databaseID
		*(values[9].(*int64)) = 2
		*(values[10].(*string)) = "trust"
		*(values[11].(**time.Time)) = &now
		*(values[12].(*time.Time)) = now
		*(values[13].(*time.Time)) = now
		return nil
	}}
	record, err := operationStore(database).GetRecoverySet(context.Background(), "set-id")
	if err != nil || record.RecoverySetID != "set-id" || record.ManifestSHA256 != "manifest" {
		t.Fatalf("get recovery set = %#v %v", record, err)
	}
}
