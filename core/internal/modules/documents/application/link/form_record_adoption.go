package link

import (
	"context"
	"errors"
	"fmt"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (registry *FormRecordRegistry) AdoptPage(
	ctx context.Context,
	scope platformtx.TenantScope,
	page formrecordlink.AdoptionPage,
) (formrecordlink.AdoptionProgress, error) {
	if registry == nil || !validAdoptionPage(page) ||
		!formRecordOperation(scope.OperationID()) ||
		platformtx.RequireOwner(scope, "forms") != nil ||
		platformtx.RequireOwner(scope, "documents") != nil {
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrInvalid
	}
	store, identity, err := registry.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrUnavailable
	}
	adoptionStore, ok := store.(AdoptionStore)
	if !ok {
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrUnavailable
	}
	return registry.adoptPage(ctx, adoptionStore, identity, page)
}

func (registry *FormRecordRegistry) adoptPage(
	ctx context.Context, store AdoptionStore, identity foundation.Identity,
	page formrecordlink.AdoptionPage,
) (formrecordlink.AdoptionProgress, error) {
	seed := adoptionSeed(page.OperationID, page.DefinitionID, page.Epoch)
	current, err := store.LoadFormRecordAdoption(ctx, seed, page.AppliedAt)
	if err != nil {
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrUnavailable
	}
	expected := adoptionCheckpoint(page.Previous)
	next, valid := adoptionPageCheckpoint(expected, page)
	if !valid {
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrInvalid
	}
	if current != expected {
		if current == next {
			return exportAdoption(current), nil
		}
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrReferenceInvalid
	}
	applied, err := registry.applyAdoptionEntries(ctx, store, identity, expected, page)
	if err != nil {
		return formrecordlink.AdoptionProgress{}, err
	}
	if applied != next {
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrUnavailable
	}
	stored, err := store.AdvanceFormRecordAdoption(
		ctx, expected, next, page.AppliedAt,
	)
	if err != nil || stored != next {
		return formrecordlink.AdoptionProgress{}, formrecordlink.ErrUnavailable
	}
	return exportAdoption(stored), nil
}

func adoptionPageCheckpoint(
	progress FormRecordAdoptionCheckpoint, page formrecordlink.AdoptionPage,
) (FormRecordAdoptionCheckpoint, bool) {
	for _, entry := range page.Entries {
		next, ok := formrecordlink.AdvanceAdoption(exportAdoption(progress), entry)
		if !ok {
			return FormRecordAdoptionCheckpoint{}, false
		}
		progress = adoptionCheckpoint(next)
	}
	progress.Done = page.Done
	return progress, true
}

func (registry *FormRecordRegistry) applyAdoptionEntries(
	ctx context.Context, store Store, identity foundation.Identity,
	progress FormRecordAdoptionCheckpoint, page formrecordlink.AdoptionPage,
) (FormRecordAdoptionCheckpoint, error) {
	for _, entry := range page.Entries {
		request := formrecordlink.Request{DefinitionID: page.DefinitionID,
			RecordID: entry.RecordID, DesiredItemIDs: entry.DesiredItemIDs,
			IdempotencyKey: fmt.Sprintf("adopt:%x:%x", page.OperationID, entry.RecordID)}
		if err := registry.reconcileFormRecord(ctx, store, identity, request); err != nil {
			if errors.Is(err, formrecordlink.ErrReferenceInvalid) {
				return FormRecordAdoptionCheckpoint{}, err
			}
			return FormRecordAdoptionCheckpoint{}, formrecordlink.ErrUnavailable
		}
	}
	next, ok := adoptionPageCheckpoint(progress, page)
	if !ok {
		return FormRecordAdoptionCheckpoint{}, formrecordlink.ErrInvalid
	}
	return next, nil
}
