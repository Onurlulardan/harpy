package publication

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type publicationMutationStore struct {
	writes   int
	create   func(CreateMutation) (Snapshot, bool, error)
	submit   func(TransitionMutation) (Snapshot, bool, error)
	approve  func(TransitionMutation) (Snapshot, bool, error)
	activate func(ActivateMutation) (Snapshot, bool, error)
}

func (store *publicationMutationStore) Create(
	_ context.Context, mutation CreateMutation,
) (Snapshot, bool, error) {
	store.writes++
	if store.create != nil {
		return store.create(mutation)
	}
	return Snapshot{}, false, nil
}

func (store *publicationMutationStore) Submit(
	_ context.Context, mutation TransitionMutation,
) (Snapshot, bool, error) {
	store.writes++
	if store.submit != nil {
		return store.submit(mutation)
	}
	return Snapshot{}, false, nil
}

func (store *publicationMutationStore) Approve(
	_ context.Context, mutation TransitionMutation,
) (Snapshot, bool, error) {
	store.writes++
	if store.approve != nil {
		return store.approve(mutation)
	}
	return Snapshot{}, false, nil
}

func (store *publicationMutationStore) Activate(
	_ context.Context, mutation ActivateMutation,
) (Snapshot, bool, error) {
	store.writes++
	if store.activate != nil {
		return store.activate(mutation)
	}
	return Snapshot{}, false, nil
}

func (*publicationMutationStore) Get(context.Context, uuid.UUID) (Snapshot, error) {
	return Snapshot{}, ErrUnavailable
}

func (*publicationMutationStore) List(context.Context, int, *uuid.UUID) (Page, error) {
	return Page{}, ErrUnavailable
}

func (store *publicationMutationStore) Revoke(
	context.Context, TransitionMutation,
) (Snapshot, bool, error) {
	store.writes++
	return Snapshot{}, false, nil
}

type publicationMutationIDs struct {
	values []uuid.UUID
	index  int
}

func (source *publicationMutationIDs) NewUUIDv7() (uuid.UUID, error) {
	if source.index >= len(source.values) {
		return uuid.Nil, ErrUnavailable
	}
	value := source.values[source.index]
	source.index++
	return value, nil
}

func publicationMutationApplication(
	entitlements *publicationFixtureEntitlements,
	ids ...uuid.UUID,
) *Application {
	return &Application{
		entitlements: entitlements,
		ids:          &publicationMutationIDs{values: ids},
		clock:        publicationFixtureClock{value: publicationFixtureNow},
	}
}

func publicationMutationCurrent() Snapshot {
	return Snapshot{
		ReleaseID:         publicationFixtureRelease,
		ResourceKind:      ResourceQuery,
		ResourceVersionID: publicationFixtureVersion,
	}
}

func publicationMutationTiming() timeBinding {
	return timeBinding{
		now:    publicationFixtureNow,
		stepUp: publicationFixtureNow.Add(-time.Minute),
	}
}

func publicationMutationIdentity() ScopeIdentity {
	return ScopeIdentity{
		ActorID:  publicationFixtureActor,
		TenantID: publicationFixtureTenant,
	}
}

var _ Store = (*publicationMutationStore)(nil)
var _ IDSource = (*publicationMutationIDs)(nil)
