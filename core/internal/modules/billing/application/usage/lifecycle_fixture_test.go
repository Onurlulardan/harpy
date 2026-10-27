package usage

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func lifecycleFixture(t *testing.T) (*Commands, *testStore, Reservation, time.Time) {
	t.Helper()
	now := time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC)
	key, err := entitlement.NewKey("iam.active_users")
	if err != nil {
		t.Fatal(err)
	}
	reservation := Reservation{
		ID: mustV7(t), DecisionID: mustV7(t), EntitlementKey: key,
		PeriodID: mustV7(t), SourceOperationID: mustV7(t), Amount: 2,
		State: usagedomain.Reserved, ExpiresAt: now.Add(time.Hour), Revision: 1,
	}
	store := &testStore{locked: reservation, lockedFound: true}
	commands, err := NewCommands(
		store,
		func() (uuid.UUID, error) { return mustV7(t), nil },
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	return commands, store, reservation, now
}

func identifierPointer(value uuid.UUID) *uuid.UUID { return &value }
func instantPointer(value time.Time) *time.Time    { return &value }
