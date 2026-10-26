package formrecordarchive

import (
	"context"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) commandScope(ctx context.Context, actorID,
	tenantID domain.ID, operation string,
) (platformtx.ScopeRequest, error) {
	tenant := uuid.UUID(tenantID)
	billingGroup, billing, forms, err := application.dependencies.Generations.
		ResolveTenantBillingGenerations(ctx, tenant)
	if err != nil {
		return platformtx.ScopeRequest{}, formrecordarchiveapi.ErrUnavailable
	}
	documentsGroup, documents, err := application.dependencies.Generations.
		ResolveTenantDocumentsGenerations(ctx, tenant)
	if err != nil || !positive(billingGroup, billing, forms, documentsGroup, documents) {
		return platformtx.ScopeRequest{}, formrecordarchiveapi.ErrUnavailable
	}
	request, err := platformtx.AdmitMultiGroupCommandScope(uuid.UUID(actorID), tenant,
		operation, []platformtx.CommandAuthorityGroup{
			{ID: "billing_forms", Generation: billingGroup},
			{ID: "documents", Generation: documentsGroup},
		}, []platformtx.CommandAuthorityDomain{
			{Name: "billing", Generation: billing},
			{Name: "documents", Generation: documents},
			{Name: "forms", Generation: forms},
		}, []string{"billing", "documents"})
	if err != nil {
		return platformtx.ScopeRequest{}, formrecordarchiveapi.ErrUnavailable
	}
	return request, nil
}

func (application *Application) snapshotScope(ctx context.Context, actorID,
	tenantID domain.ID, operation string,
) (platformtx.ScopeRequest, error) {
	_, generation, err := application.dependencies.Generations.
		ResolveTenantDocumentsGenerations(ctx, uuid.UUID(tenantID))
	if err != nil || generation < 1 {
		return platformtx.ScopeRequest{}, formrecordarchiveapi.ErrUnavailable
	}
	actor, err := actorcontext.Issue(uuid.UUID(actorID), uuid.UUID(tenantID))
	if err != nil {
		return platformtx.ScopeRequest{}, formrecordarchiveapi.ErrUnavailable
	}
	operationAuthority, err := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"documents": generation})
	if err != nil {
		return platformtx.ScopeRequest{}, formrecordarchiveapi.ErrUnavailable
	}
	return platformtx.ScopeRequest{Actor: actor, Operation: operationAuthority,
		Mode: platformtx.SnapshotQuery}, nil
}

func positive(values ...int64) bool {
	for _, value := range values {
		if value < 1 {
			return false
		}
	}
	return true
}
