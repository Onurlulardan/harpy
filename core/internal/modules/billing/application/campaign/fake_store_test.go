package campaign

import (
	"context"
	"testing"
	"time"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
	"github.com/google/uuid"
)

type fakeStore struct {
	create        func(CreateMutation) (Snapshot, error)
	createVersion func(CreateVersionMutation) (Snapshot, error)
	retire        func(RetireMutation) (Snapshot, error)
}

func (store *fakeStore) Create(
	_ context.Context,
	mutation CreateMutation,
) (Snapshot, error) {
	if store.create == nil {
		panic("unexpected campaign create")
	}
	return store.create(mutation)
}

func (store *fakeStore) Publish(
	_ context.Context,
	mutation CreateVersionMutation,
) (Snapshot, error) {
	if store.createVersion == nil {
		panic("unexpected campaign create version")
	}
	return store.createVersion(mutation)
}

func (store *fakeStore) Retire(
	_ context.Context,
	mutation RetireMutation,
) (Snapshot, error) {
	if store.retire == nil {
		panic("unexpected campaign retire")
	}
	return store.retire(mutation)
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

func publicationFixture(
	t *testing.T,
	identifiers ...campaigndomain.Identifier,
) campaigndomain.Publication {
	t.Helper()
	publication, err := campaigndomain.NewPublication(
		"Launch 2026", "Loyal customers", "Early access", true, 7,
		time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC), nil, identifiers,
	)
	if err != nil {
		t.Fatal(err)
	}
	return publication
}

func campaignKey(t *testing.T) campaigndomain.Key {
	t.Helper()
	key, err := campaigndomain.NewKey("launch-2026")
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func discountID(last byte) campaigndomain.Identifier {
	var identifier campaigndomain.Identifier
	identifier[15] = last
	return identifier
}

func v7(t *testing.T) uuid.UUID {
	t.Helper()
	identifier, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return identifier
}
