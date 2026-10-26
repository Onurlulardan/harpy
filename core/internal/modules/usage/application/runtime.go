package application

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type Service struct {
	store Store
	clock Clock
	ids   IDGenerator
}

func NewService(store Store, clock Clock, ids IDGenerator) (*Service, error) {
	if store == nil || clock == nil || ids == nil {
		return nil, fmt.Errorf("usage service dependencies are required")
	}
	return &Service{store: store, clock: clock, ids: ids}, nil
}

func (service *Service) ready(ctx context.Context) error {
	if service == nil || service.store == nil || service.clock == nil ||
		service.ids == nil || ctx == nil || ctx.Err() != nil {
		return domain.ErrUnavailable
	}
	return nil
}

func authorize(access Access, tenantOperation, platformOperation string) error {
	operation := tenantOperation
	if access.Scope() == domain.PlatformScope {
		operation = platformOperation
	}
	if !access.Allows(access.Scope(), operation) {
		return domain.ErrForbidden
	}
	return nil
}
