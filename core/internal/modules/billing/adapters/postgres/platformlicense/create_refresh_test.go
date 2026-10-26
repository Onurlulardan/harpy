package platformlicense

import (
	"testing"
	"time"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

func TestCreateUsesDurableRefreshRoutine(t *testing.T) {
	operationID := testV7(t)
	createdAt := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	query := &queryFixture{scan: func(destinations ...any) error {
		*destinations[0].(*uuid.UUID) = operationID
		*destinations[1].(*bool) = true
		*destinations[2].(*time.Time) = createdAt
		return nil
	}}
	command := platformlicenseapp.RefreshCommand{
		EnvironmentID: testV7(t), TenantID: testV7(t), ActorID: testV7(t),
		IdempotencyKey: "license-refresh-0001",
		RequestSHA256:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	value, err := fixtureStore(t, query).Create(t.Context(), command)
	if err != nil || value.OperationID != operationID || !value.Replayed {
		t.Fatalf("refresh=%+v err=%v", value, err)
	}
	if query.statement != requestRefreshSQL || len(query.arguments) != 5 ||
		query.arguments[3] != command.IdempotencyKey {
		t.Fatalf("statement or arguments mismatch")
	}
}
