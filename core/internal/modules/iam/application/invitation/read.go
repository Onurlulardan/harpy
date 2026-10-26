package invitation

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ReadQuery struct{ ID uuid.UUID }

func (query ReadQuery) Validate() error {
	if query.ID == uuid.Nil {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Read(
	ctx context.Context,
	scope platformtx.TenantScope,
	query ReadQuery,
) (Invitation, error) {
	if service == nil || service.store == nil || query.Validate() != nil ||
		scope.ActorID() == uuid.Nil {
		return Invitation{}, ErrInvalid
	}
	result, err := service.store.Read(ctx, scope, query)
	if err != nil {
		return Invitation{}, err
	}
	if result.Validate() != nil {
		return Invitation{}, ErrUnavailable
	}
	return result, nil
}
