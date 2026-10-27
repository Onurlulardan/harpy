package offering

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
	"github.com/google/uuid"
)

type storeFixture struct {
	create        func(CreateMutation) (Snapshot, error)
	createVersion func(CreateVersionMutation) (Snapshot, error)
	retire        func(RetireMutation) (Snapshot, error)
}

func (store storeFixture) Create(_ context.Context, value CreateMutation) (Snapshot, error) {
	return store.create(value)
}

func (store storeFixture) Publish(
	_ context.Context, value CreateVersionMutation,
) (Snapshot, error) {
	return store.createVersion(value)
}

func (store storeFixture) Retire(_ context.Context, value RetireMutation) (Snapshot, error) {
	return store.retire(value)
}

func commandFixture(t *testing.T, store Store) *Commands {
	t.Helper()
	commands, err := NewCommands(store, uuid.NewV7, func() time.Time {
		return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	})
	if err != nil {
		t.Fatal(err)
	}
	return commands
}

func versionFixture(t *testing.T) offeringdomain.Version {
	t.Helper()
	amount, _ := money.New(9900, "USD", money.RegistryVersion)
	price, _ := offeringdomain.NewPrice(offeringdomain.Monthly, amount)
	entry := entitlement.RegistryV1ByKey()["iam.active_users"]
	limit, _ := entitlement.IntegerValue(500)
	grant, _ := offeringdomain.NewGrant(entry.Definition, limit)
	value, err := offeringdomain.NewVersion(
		"Pro", true, true, false, 14, 3,
		time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC), nil,
		[]offeringdomain.Price{price}, []offeringdomain.Grant{grant},
	)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func v7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
