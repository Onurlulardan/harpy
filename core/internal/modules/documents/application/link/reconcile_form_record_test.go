package link

import (
	"context"
	"errors"
	"testing"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestReconcileFormRecordDiffsRecordWideDesiredSet(t *testing.T) {
	first := domain.ID(formRecordUUID(21))
	second := domain.ID(formRecordUUID(22))
	third := domain.ID(formRecordUUID(23))
	store := newFormRecordStoreFixture(first, second, third)
	registry, identity := formRecordRegistryFixture()
	request := formRecordRequest("create-0001", first, first, second)

	if err := registry.reconcileFormRecord(
		context.Background(), store, identity, request,
	); err != nil {
		t.Fatal(err)
	}
	if store.registerCalls != 2 || len(store.active) != 2 {
		t.Fatalf("create calls=%d active=%d", store.registerCalls, len(store.active))
	}
	if store.active[first].ReferenceEpoch != store.epochs[first] ||
		store.active[second].ReferenceEpoch != store.epochs[second] {
		t.Fatal("link epoch did not come from the Documents store")
	}
	keptID := store.active[first].ID

	request = formRecordRequest("update-0001", first, third, first)
	if err := registry.reconcileFormRecord(
		context.Background(), store, identity, request,
	); err != nil {
		t.Fatal(err)
	}
	if store.registerCalls != 3 || store.releaseCalls != 1 ||
		len(store.active) != 2 || store.active[first].ID != keptID {
		t.Fatalf("update calls register=%d release=%d active=%d",
			store.registerCalls, store.releaseCalls, len(store.active))
	}
	if _, exists := store.active[second]; exists {
		t.Fatal("removed item remains active")
	}

	if err := registry.reconcileFormRecord(
		context.Background(), store, identity, request,
	); err != nil {
		t.Fatal(err)
	}
	if store.registerCalls != 3 || store.releaseCalls != 1 {
		t.Fatal("desired-set replay performed mutations")
	}
}

func TestReconcileFormRecordDeleteThenReaddCreatesNewLink(t *testing.T) {
	itemID := domain.ID(formRecordUUID(31))
	store := newFormRecordStoreFixture(itemID)
	registry, identity := formRecordRegistryFixture()
	create := formRecordRequest("create-0002", itemID)
	if err := registry.reconcileFormRecord(
		context.Background(), store, identity, create,
	); err != nil {
		t.Fatal(err)
	}
	firstLinkID := store.active[itemID].ID
	if err := registry.reconcileFormRecord(
		context.Background(), store, identity, formRecordRequest("delete-0002"),
	); err != nil {
		t.Fatal(err)
	}
	if len(store.active) != 0 || len(store.released) != 1 {
		t.Fatal("delete did not release every active link")
	}
	if err := registry.reconcileFormRecord(
		context.Background(), store, identity,
		formRecordRequest("readd-0002", itemID),
	); err != nil {
		t.Fatal(err)
	}
	if store.active[itemID].ID == firstLinkID ||
		store.released[0].ID != firstLinkID {
		t.Fatal("re-add reused the immutable released link")
	}
}

func TestReconcileFormRecordMapsMissingItemAndStoreFailure(t *testing.T) {
	itemID := domain.ID(formRecordUUID(41))
	registry, identity := formRecordRegistryFixture()
	store := newFormRecordStoreFixture()
	err := registry.reconcileFormRecord(
		context.Background(), store, identity,
		formRecordRequest("missing-0001", itemID),
	)
	if !errors.Is(err, formrecordlink.ErrReferenceInvalid) {
		t.Fatalf("missing item error = %v", err)
	}
	store.listErr = ErrUnavailable
	err = registry.reconcileFormRecord(
		context.Background(), store, identity,
		formRecordRequest("failure-0001"),
	)
	if !errors.Is(err, formrecordlink.ErrUnavailable) {
		t.Fatalf("store failure error = %v", err)
	}
}

func formRecordRequest(
	key string,
	items ...domain.ID,
) formrecordlink.Request {
	result := formrecordlink.Request{
		DefinitionID:   formrecordlink.ID(formRecordUUID(11)),
		RecordID:       formrecordlink.ID(formRecordUUID(12)),
		IdempotencyKey: key,
	}
	for _, itemID := range items {
		result.DesiredItemIDs = append(
			result.DesiredItemIDs, formrecordlink.ID(itemID),
		)
	}
	return result
}
