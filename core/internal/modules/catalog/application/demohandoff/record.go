package demohandoff

import (
	"context"

	api "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/demohandoff"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Projection struct {
	ProposedHandoffID uuid.UUID
	Request           api.Request
	OwnerGeneration   int64
}

func (application *Application) RecordDemoHandoff(
	ctx context.Context,
	scope platformtx.TenantScope,
	request api.Request,
) (api.Receipt, error) {
	if application == nil || !requestValid(request) || !scopeValid(ctx, scope) {
		return api.Receipt{}, api.ErrInvalid
	}
	authority, err := platformtx.MintOwnerAuthority(scope, owner)
	if err != nil || authority.Generation() < 1 {
		return api.Receipt{}, api.ErrUnavailable
	}
	store, err := application.binder(ctx, scope, authority)
	if err != nil || nilDependency(store) {
		return api.Receipt{}, api.ErrUnavailable
	}
	handoffID, err := application.ids.NewUUIDv7()
	if err != nil || handoffID.Version() != 7 {
		return api.Receipt{}, api.ErrUnavailable
	}
	receipt, err := store.Persist(ctx, Projection{
		ProposedHandoffID: handoffID,
		Request:           request,
		OwnerGeneration:   authority.Generation(),
	})
	if err != nil {
		return api.Receipt{}, err
	}
	if !receiptMatches(receipt, request) {
		return api.Receipt{}, api.ErrConflict
	}
	return receipt, nil
}
