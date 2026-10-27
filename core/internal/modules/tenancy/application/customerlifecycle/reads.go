package customerlifecycle

import (
	"context"

	"github.com/google/uuid"
)

type ListQuery struct {
	Credential CredentialEnvelope
	Limit      int
	After      uuid.UUID
}

func (service *Service) List(ctx context.Context, query ListQuery) (Page, error) {
	if ctx == nil || ctx.Err() != nil || query.Limit < 0 || query.Limit > 100 {
		return Page{}, ErrInvalid
	}
	edge, err := service.boundary()
	if err != nil {
		return Page{}, err
	}
	if _, err = service.authorize(ctx, edge, query.Credential, CapabilityRead,
		uuid.Nil, false, [32]byte{}); err != nil {
		return Page{}, accessError(err)
	}
	limit := query.Limit
	if limit == 0 {
		limit = 50
	}
	page, err := service.dependencies.Store.List(ctx, ListStoreRequest{
		EnvironmentID: edge.environmentID, Limit: limit, After: query.After,
	})
	if err != nil {
		return Page{}, ErrUnavailable
	}
	if page.Items == nil {
		page.Items = []Tenant{}
	}
	return page, nil
}

type GetQuery struct {
	Credential CredentialEnvelope
	TenantID   uuid.UUID
}

func (service *Service) Get(ctx context.Context, query GetQuery) (Tenant, error) {
	if ctx == nil || ctx.Err() != nil || query.TenantID == uuid.Nil {
		return Tenant{}, ErrInvalid
	}
	edge, err := service.boundary()
	if err != nil {
		return Tenant{}, err
	}
	if query.TenantID == edge.controlTenantID {
		return Tenant{}, ErrNotFound
	}
	if _, err = service.authorize(ctx, edge, query.Credential, CapabilityRead,
		query.TenantID, false, [32]byte{}); err != nil {
		return Tenant{}, accessError(err)
	}
	result, err := service.dependencies.Store.Get(ctx, edge.environmentID, query.TenantID)
	if err != nil {
		return Tenant{}, err
	}
	return result, nil
}

type OperationQuery struct {
	Credential  CredentialEnvelope
	OperationID uuid.UUID
}

func (service *Service) ReadOperation(ctx context.Context, query OperationQuery) (Operation, error) {
	if ctx == nil || ctx.Err() != nil || query.OperationID == uuid.Nil {
		return Operation{}, ErrInvalid
	}
	edge, err := service.boundary()
	if err != nil {
		return Operation{}, err
	}
	if _, err = service.authorize(ctx, edge, query.Credential, CapabilityRead,
		uuid.Nil, false, [32]byte{}); err != nil {
		return Operation{}, accessError(err)
	}
	return service.dependencies.Store.ReadOperation(ctx, edge.controlTenantID, query.OperationID)
}
