package licenseclock

import (
	"context"
	"errors"
	"testing"
	"time"
)

type clockFixture struct {
	now time.Time
	err error
}

func (fixture clockFixture) ReadDatabaseClock(context.Context) (time.Time, error) {
	return fixture.now, fixture.err
}

func TestReadNormalizesUTC(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.FixedZone("fixture", 3*60*60))
	result, err := Read(t.Context(), clockFixture{now: now})
	if err != nil || result.Location() != time.UTC || !result.Equal(now) {
		t.Fatalf("clock=%v err=%v", result, err)
	}
}

func TestReadRejectsUnavailableAuthority(t *testing.T) {
	dependencyErr := errors.New("database unavailable")
	if _, err := Read(t.Context(), clockFixture{err: dependencyErr}); !errors.Is(
		err, dependencyErr,
	) {
		t.Fatalf("dependency error=%v", err)
	}
	if _, err := Read(t.Context(), clockFixture{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("zero clock error=%v", err)
	}
}
