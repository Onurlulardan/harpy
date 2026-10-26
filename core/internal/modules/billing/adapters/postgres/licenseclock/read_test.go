package licenseclock

import (
	"context"
	"testing"
	"time"
)

type rowFixture func(...any) error

func (fixture rowFixture) Scan(destinations ...any) error {
	return fixture(destinations...)
}

func TestReadDatabaseClockReturnsUTCValue(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.FixedZone("fixture", 3*60*60))
	store, err := NewStore(DatabaseAccess{QueryRow: func(
		context.Context, string, ...any,
	) Row {
		return rowFixture(func(destinations ...any) error {
			*destinations[0].(*time.Time) = now
			return nil
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.ReadDatabaseClock(t.Context())
	if err != nil || result.Location() != time.UTC || !result.Equal(now) {
		t.Fatalf("clock=%v err=%v", result, err)
	}
}
