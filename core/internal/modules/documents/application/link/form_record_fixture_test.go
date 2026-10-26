package link

import (
	"context"
	"slices"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

type formRecordStoreFixture struct {
	active        map[domain.ID]domain.Link
	epochs        map[domain.ID]uint64
	released      []domain.Link
	registerCalls int
	releaseCalls  int
	listErr       error
}

func newFormRecordStoreFixture(items ...domain.ID) *formRecordStoreFixture {
	fixture := &formRecordStoreFixture{
		active: make(map[domain.ID]domain.Link),
		epochs: make(map[domain.ID]uint64),
	}
	for index, itemID := range items {
		fixture.epochs[itemID] = uint64(index + 7)
	}
	return fixture
}

func (fixture *formRecordStoreFixture) Register(
	_ context.Context,
	mutation Mutation,
) (domain.Link, bool, error) {
	fixture.registerCalls++
	if _, exists := fixture.active[mutation.Link.ItemID]; exists {
		return domain.Link{}, false, ErrConflict
	}
	fixture.active[mutation.Link.ItemID] = mutation.Link
	return mutation.Link, false, nil
}

func (fixture *formRecordStoreFixture) Release(
	_ context.Context,
	mutation Mutation,
) (domain.Link, bool, error) {
	fixture.releaseCalls++
	for itemID, value := range fixture.active {
		if value.ID != mutation.Link.ID {
			continue
		}
		released, err := value.Release(mutation.OccurredAt)
		if err != nil {
			return domain.Link{}, false, err
		}
		delete(fixture.active, itemID)
		fixture.released = append(fixture.released, released)
		return released, false, nil
	}
	return domain.Link{}, false, ErrConflict
}

func (fixture *formRecordStoreFixture) SourceLinks(
	context.Context,
	SourceQuery,
) ([]domain.Link, error) {
	if fixture.listErr != nil {
		return nil, fixture.listErr
	}
	result := make([]domain.Link, 0, len(fixture.active))
	for _, value := range fixture.active {
		result = append(result, value)
	}
	slices.SortFunc(result, func(left, right domain.Link) int {
		return slices.Compare(left.ItemID[:], right.ItemID[:])
	})
	return result, nil
}

func (fixture *formRecordStoreFixture) LockActiveEpoch(
	_ context.Context,
	itemID domain.ID,
) (uint64, error) {
	value, found := fixture.epochs[itemID]
	if !found {
		return 0, ErrNotFound
	}
	return value, nil
}

func (*formRecordStoreFixture) ListBlockers(
	context.Context,
	BlockerQuery,
) ([]domain.Link, error) {
	return nil, nil
}

type formRecordIDFixture struct{ next byte }

func (fixture *formRecordIDFixture) NewUUIDv7() (domain.ID, error) {
	fixture.next++
	return domain.ID(formRecordUUID(100 + fixture.next)), nil
}

type formRecordClockFixture struct{}

func (formRecordClockFixture) Now() time.Time {
	return time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
}

func formRecordRegistryFixture() (*FormRecordRegistry, foundation.Identity) {
	return &FormRecordRegistry{
			ids: &formRecordIDFixture{}, clock: formRecordClockFixture{},
		}, foundation.Identity{
			ActorID:  domain.ID(formRecordUUID(1)),
			TenantID: domain.ID(formRecordUUID(2)),
		}
}

func formRecordUUID(seed byte) uuid.UUID {
	value := uuid.MustParse("019c7100-0000-7000-8000-000000000000")
	value[15] = seed
	return value
}
