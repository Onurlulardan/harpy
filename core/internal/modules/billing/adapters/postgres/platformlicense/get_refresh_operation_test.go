package platformlicense

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetDecodesRefreshOutcome(t *testing.T) {
	environmentID, operationID := testV7(t), testV7(t)
	sequence, digest, failure := int64(7), "digest", "outcome_unknown"
	createdAt := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	completedAt := createdAt.Add(time.Minute)
	query := &queryFixture{scan: func(destinations ...any) error {
		*destinations[0].(*uuid.UUID) = operationID
		*destinations[1].(*string) = "failed"
		*destinations[2].(**int64) = &sequence
		*destinations[3].(**string) = &digest
		*destinations[4].(**string) = &failure
		*destinations[5].(*time.Time) = createdAt
		*destinations[6].(**time.Time) = &completedAt
		return nil
	}}
	value, err := fixtureStore(t, query).Get(
		t.Context(), environmentID, operationID,
	)
	if err != nil || value.State != "failed" ||
		value.FailureCode == nil || *value.FailureCode != failure ||
		value.CompletedAt == nil || value.CompletedAt.Location() != time.UTC {
		t.Fatalf("operation=%+v err=%v", value, err)
	}
	if query.statement != readRefreshOperationSQL || len(query.arguments) != 2 {
		t.Fatalf("statement or arguments mismatch")
	}
}
