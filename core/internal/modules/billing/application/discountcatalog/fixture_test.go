package discountcatalog

import (
	"context"
	"testing"
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
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

func commandsFixture(t *testing.T, store Store) *Commands {
	t.Helper()
	commands, err := NewCommands(store, uuid.NewV7, func() time.Time {
		return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	})
	if err != nil {
		t.Fatal(err)
	}
	return commands
}

func policyFixture(t *testing.T) discountdomain.Policy {
	t.Helper()
	code, _ := discountdomain.NormalizeCode("SAVE10")
	rate, _ := money.NewBasisPoints(1000)
	policy, err := discountdomain.NewBasisPointPolicy(
		code, rate, discountdomain.OnceDuration(), 1, nil, nil,
		time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC), nil, true, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func presentationFixture(t *testing.T) discountdomain.Presentation {
	t.Helper()
	presentation, err := discountdomain.NewPresentation("Save ten", "Loyal customers")
	if err != nil {
		t.Fatal(err)
	}
	return presentation
}

func v7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
