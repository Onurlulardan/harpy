package link

import (
	"context"
	"errors"
	"testing"
	"time"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestReconcileFormRecordAdoptionUsesRecordWideDesiredSet(t *testing.T) {
	itemA := domain.ID(formRecordUUID(61))
	itemB := domain.ID(formRecordUUID(62))
	store := newFormRecordStoreFixture(itemA, itemB)
	registry, identity := formRecordRegistryFixture()
	progress := formRecordAdoptionProgress()
	page := formrecordlink.AdoptionPage{OperationID: progress.OperationID,
		DefinitionID: progress.DefinitionID, Epoch: progress.Epoch,
		Previous: progress, Done: true,
		AppliedAt: time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Entries: []formrecordlink.AdoptionEntry{{
			RecordID:     formrecordlink.ID(formRecordUUID(63)),
			SourceSHA256: [32]byte{7}, DesiredItemIDs: []formrecordlink.ID{
				formrecordlink.ID(itemB), formrecordlink.ID(itemA),
			},
		}},
	}
	next, err := registry.applyAdoptionEntries(
		context.Background(), store, identity, adoptionCheckpoint(progress), page,
	)
	if err != nil || !next.Done || next.RecordCount != 1 ||
		next.LinkCount != 2 || len(store.active) != 2 {
		t.Fatalf("adoption next=%+v active=%d err=%v", next, len(store.active), err)
	}
}

func TestReconcileFormRecordAdoptionDeletedRowReleasesExistingLink(t *testing.T) {
	itemID := domain.ID(formRecordUUID(71))
	store := newFormRecordStoreFixture(itemID)
	registry, identity := formRecordRegistryFixture()
	progress := formRecordAdoptionProgress()
	setup := formRecordRequest("adoption-delete-setup", itemID)
	setup.DefinitionID = progress.DefinitionID
	setup.RecordID = formrecordlink.ID(formRecordUUID(72))
	if err := registry.reconcileFormRecord(
		context.Background(), store, identity, setup,
	); err != nil {
		t.Fatal(err)
	}
	page := formrecordlink.AdoptionPage{OperationID: progress.OperationID,
		DefinitionID: progress.DefinitionID, Epoch: progress.Epoch,
		Previous: progress, Done: true,
		AppliedAt: time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Entries: []formrecordlink.AdoptionEntry{{
			RecordID:     formrecordlink.ID(formRecordUUID(72)),
			SourceSHA256: [32]byte{8},
		}},
	}
	if _, err := registry.applyAdoptionEntries(
		context.Background(), store, identity, adoptionCheckpoint(progress), page,
	); err != nil || len(store.active) != 0 || store.releaseCalls != 1 {
		t.Fatalf("deleted adoption active=%d releases=%d err=%v",
			len(store.active), store.releaseCalls, err)
	}
}

func TestAdoptPageDoesNotMutateBeforeCheckpointMatch(t *testing.T) {
	itemID := domain.ID(formRecordUUID(81))
	base := newFormRecordStoreFixture(itemID)
	registry, identity := formRecordRegistryFixture()
	progress := formRecordAdoptionProgress()
	page := formrecordlink.AdoptionPage{OperationID: progress.OperationID,
		DefinitionID: progress.DefinitionID, Epoch: progress.Epoch,
		Previous: progress, AppliedAt: time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Entries: []formrecordlink.AdoptionEntry{{
			RecordID: formrecordlink.ID(formRecordUUID(82)), SourceSHA256: [32]byte{9},
			DesiredItemIDs: []formrecordlink.ID{formrecordlink.ID(itemID)},
		}},
	}
	next, ok := adoptionPageCheckpoint(adoptionCheckpoint(progress), page)
	if !ok {
		t.Fatal("valid adoption page did not advance")
	}

	replay := &adoptionStoreFixture{formRecordStoreFixture: base, checkpoint: next}
	if result, err := registry.adoptPage(
		context.Background(), replay, identity, page,
	); err != nil || adoptionCheckpoint(result) != next || base.registerCalls != 0 {
		t.Fatalf("replay result=%+v calls=%d err=%v", result, base.registerCalls, err)
	}

	mismatch := next
	mismatch.SourceSHA256[0] ^= 0xff
	replay.checkpoint = mismatch
	if _, err := registry.adoptPage(
		context.Background(), replay, identity, page,
	); !errors.Is(err, formrecordlink.ErrReferenceInvalid) || base.registerCalls != 0 {
		t.Fatalf("mismatch calls=%d err=%v", base.registerCalls, err)
	}
}

type adoptionStoreFixture struct {
	*formRecordStoreFixture
	checkpoint FormRecordAdoptionCheckpoint
}

func (fixture *adoptionStoreFixture) LoadFormRecordAdoption(
	context.Context, FormRecordAdoptionCheckpoint, time.Time,
) (FormRecordAdoptionCheckpoint, error) {
	return fixture.checkpoint, nil
}

func (fixture *adoptionStoreFixture) AdvanceFormRecordAdoption(
	_ context.Context, _ FormRecordAdoptionCheckpoint,
	next FormRecordAdoptionCheckpoint, _ time.Time,
) (FormRecordAdoptionCheckpoint, error) {
	fixture.checkpoint = next
	return next, nil
}

func (*adoptionStoreFixture) SealFormRecordAdoption(
	context.Context, FormRecordAdoptionSeal,
) (FormRecordAdoptionReceipt, error) {
	return FormRecordAdoptionReceipt{}, nil
}

func formRecordAdoptionProgress() formrecordlink.AdoptionProgress {
	operationID := formrecordlink.ID(formRecordUUID(51))
	definitionID := formrecordlink.ID(formRecordUUID(52))
	source, catalog := formrecordlink.AdoptionSeeds(definitionID, 2)
	return formrecordlink.AdoptionProgress{OperationID: operationID,
		DefinitionID: definitionID, Epoch: 2,
		SourceSHA256: source, LinkSHA256: catalog}
}
