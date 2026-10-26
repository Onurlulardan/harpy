package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ListQuery struct {
	States []State
	Limit  int
	After  uuid.UUID
}

type Page struct {
	Groups []Group
	Next   uuid.UUID
}

func (query ListQuery) Validate() error {
	if query.Limit < 1 || query.Limit > 100 || len(query.States) > 6 {
		return ErrInvalid
	}
	seen := make(map[State]struct{}, len(query.States))
	for _, state := range query.States {
		if !state.Valid() {
			return ErrInvalid
		}
		if _, exists := seen[state]; exists {
			return ErrInvalid
		}
		seen[state] = struct{}{}
	}
	return nil
}

func (service *Service) List(
	ctx context.Context,
	scope transaction.TenantScope,
	query ListQuery,
) (Page, error) {
	if service == nil || service.store == nil || query.Validate() != nil ||
		scope.ActorID() == uuid.Nil {
		return Page{}, ErrInvalid
	}
	page, err := service.store.List(ctx, scope, query)
	if err != nil {
		return Page{}, err
	}
	for _, group := range page.Groups {
		if group.Validate() != nil {
			return Page{}, ErrUnavailable
		}
	}
	return page, nil
}
