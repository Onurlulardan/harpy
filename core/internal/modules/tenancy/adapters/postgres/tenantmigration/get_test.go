package tenantmigration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type rowFixture func(...any) error

func (fixture rowFixture) Scan(destinations ...any) error {
	return fixture(destinations...)
}

func TestGetUsesNarrowOwnerRoutine(t *testing.T) {
	operationID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	repositoryPort, err := NewRepository(func(
		_ context.Context, statement string, arguments ...any,
	) Row {
		if len(arguments) != 1 || arguments[0] != operationID ||
			!strings.Contains(statement, "tenancy_tenant_migration_get_v2") {
			t.Fatalf("query = %q, %#v", statement, arguments)
		}
		return rowFixture(func(destinations ...any) error {
			*destinations[0].(*string) = "22222222-2222-4222-8222-222222222222"
			*destinations[1].(*string) = operationID.String()
			*destinations[2].(*string) = "running"
			*destinations[3].(*string) = "33333333-3333-4333-8333-333333333333"
			*destinations[4].(*string) = "t_33333333333343338333333333333333"
			*destinations[5].(*string) = strings.Repeat("a", 64)
			*destinations[6].(*string) = "44444444-4444-4444-8444-444444444444"
			*destinations[7].(*int64) = 13
			*destinations[8].(*int64) = 15
			*destinations[9].(*string) = strings.Repeat("b", 64)
			*destinations[15].(*time.Time) = time.Date(2026, 7, 22, 11, 59, 0, 0, time.UTC)
			*destinations[16].(*time.Time) = time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
			*destinations[17].(*bool) = true
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	job, found, err := repositoryPort.Get(context.Background(), operationID)
	if err != nil || !found || job.OperationID != operationID.String() {
		t.Fatalf("Get() = %#v, %t, %v", job, found, err)
	}
}
