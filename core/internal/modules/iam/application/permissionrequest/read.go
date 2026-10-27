package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ReadQuery struct {
	GroupID uuid.UUID
}

func (query ReadQuery) Validate() error {
	if query.GroupID == uuid.Nil {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Read(
	ctx context.Context,
	scope transaction.TenantScope,
	query ReadQuery,
) (Group, error) {
	if service == nil || service.store == nil || query.Validate() != nil ||
		scope.ActorID() == uuid.Nil {
		return Group{}, ErrInvalid
	}
	result, err := service.store.Read(ctx, scope, query)
	if err != nil {
		return Group{}, err
	}
	if result.Validate() != nil {
		return Group{}, ErrUnavailable
	}
	return result, nil
}
