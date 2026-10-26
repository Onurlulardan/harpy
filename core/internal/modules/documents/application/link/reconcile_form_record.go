package link

import (
	"context"
	"errors"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (registry *FormRecordRegistry) Reconcile(
	ctx context.Context,
	scope platformtx.TenantScope,
	request formrecordlink.Request,
) error {
	if registry == nil || !request.Valid() || !formRecordOperation(scope.OperationID()) ||
		platformtx.RequireOwner(scope, "forms") != nil ||
		platformtx.RequireOwner(scope, "documents") != nil {
		return formrecordlink.ErrInvalid
	}
	store, identity, err := registry.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return formrecordlink.ErrUnavailable
	}
	return registry.reconcileFormRecord(ctx, store, identity, request)
}

func (registry *FormRecordRegistry) reconcileFormRecord(
	ctx context.Context,
	store Store,
	identity foundation.Identity,
	request formrecordlink.Request,
) error {
	if registry == nil || ctx == nil || ctx.Err() != nil || store == nil ||
		!identity.Valid() || !request.Valid() {
		return formrecordlink.ErrInvalid
	}
	source, desired, err := canonicalFormRecordRequest(request)
	if err != nil {
		return err
	}
	current, err := listFormRecordLinks(ctx, store, source)
	if err != nil || len(current) > formrecordlink.MaximumItemCount {
		return formrecordlink.ErrUnavailable
	}
	mutations, err := formRecordDiff(identity.TenantID, source, current, desired)
	if err != nil {
		return err
	}
	now := registry.clock.Now()
	if !foundation.ValidInstant(now) {
		return formrecordlink.ErrUnavailable
	}
	for _, mutation := range mutations {
		if mutation.release != nil {
			err = registry.releaseFormRecordLink(
				ctx, store, identity, source, *mutation.release, request, now,
			)
		} else {
			err = registry.registerFormRecordLink(
				ctx, store, identity, source, mutation.itemID, request, now,
			)
		}
		if err != nil {
			if errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) ||
				errors.Is(err, ErrInvalid) {
				return formrecordlink.ErrReferenceInvalid
			}
			return formrecordlink.ErrUnavailable
		}
	}
	return nil
}

func formRecordOperation(value string) bool {
	switch value {
	case "forms.record_kernel.create", "forms.record_kernel.update",
		"forms.record_kernel.delete", "forms.record_import.execute_batch",
		"workflow.runtime.execute.forms", "catalog.operation.resume":
		return true
	default:
		return false
	}
}
