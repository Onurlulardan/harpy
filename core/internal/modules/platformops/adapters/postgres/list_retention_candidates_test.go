package postgres

import (
	"context"
	"testing"
	"time"
)

func TestListRetentionCandidates(t *testing.T) {
	now := time.Date(2026, 7, 23, 1, 0, 0, 0, time.UTC)
	rows := &operationRows{rows: []func(...any) error{func(values ...any) error {
		*(values[0].(*string)) = "set-id"
		*(values[1].(*string)) = "manifest"
		*(values[2].(*time.Time)) = now
		*(values[3].(*bool)) = true
		return nil
	}}}
	records, err := operationStore(&operationDatabase{rows: rows}).
		ListRetentionCandidates(context.Background(), "environment-id")
	if err != nil || len(records) != 1 || !records[0].LegalHold {
		t.Fatalf("retention records = %#v %v", records, err)
	}
}
