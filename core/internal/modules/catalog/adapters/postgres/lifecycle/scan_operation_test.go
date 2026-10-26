package lifecycle

import (
	"testing"
	"time"
)

func TestOperationRecordNormalizesDatabaseInstantsToUTC(t *testing.T) {
	location := time.FixedZone("database-session", 3*60*60)
	started := time.Date(2026, 7, 18, 21, 45, 0, 123000, location)
	destructive := started.Add(time.Minute)
	record := operationRecord{
		StartedAt: started, DestructiveStartedAt: &destructive,
	}

	record.normalizeTimes()

	if record.StartedAt.Location() != time.UTC ||
		record.DestructiveStartedAt == nil ||
		record.DestructiveStartedAt.Location() != time.UTC {
		t.Fatal("database instants were not normalized to UTC")
	}
	if !record.StartedAt.Equal(started) || !record.DestructiveStartedAt.Equal(destructive) {
		t.Fatal("UTC normalization changed the represented instant")
	}
}
