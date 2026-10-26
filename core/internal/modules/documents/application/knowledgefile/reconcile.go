package knowledgefile

import (
	"context"
	"fmt"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ReconcileExpired(ctx context.Context,
	request knowledgeapi.ReconcileRequest,
) (knowledgeapi.Reconciled, error) {
	access := Access{TenantID: domain.ID(request.TenantID), ActorID: domain.ID(request.ActorID),
		AuthorizationVersion: request.AuthorizationVersion}
	if application == nil || !access.TenantID.Valid() || !access.ActorID.Valid() ||
		access.AuthorizationVersion < 1 ||
		request.Limit < 1 || request.Limit > 100 {
		return knowledgeapi.Reconciled{}, knowledgeapi.ErrInvalid
	}
	result := knowledgeapi.Reconciled{}
	err := application.scopes.InCommand(ctx, access, itemapp.OperationDelete,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			leases, identity, bindErr := application.leases.Bind(scoped, scope)
			if bindErr != nil || identity.TenantID != access.TenantID ||
				identity.ActorID != access.ActorID {
				return errUnavailable
			}
			expired, expireErr := leases.Expire(scoped,
				application.clock.Now().UTC(), request.Limit)
			if expireErr != nil {
				return expireErr
			}
			for _, lease := range expired {
				deleted, deleteErr := application.items.Delete(scoped, scope,
					itemapp.MutationRequest{ItemID: lease.ItemID,
						RowVersion: lease.ItemRowVersion,
						IdempotencyKey: fmt.Sprintf("ai-knowledge-expire-%s",
							lease.LeaseID.String())})
				if deleteErr != nil || deleted.ID != lease.ItemID {
					return errUnavailable
				}
				result.ExpiredItemIDs = append(result.ExpiredItemIDs,
					knowledgeapi.ID(lease.ItemID))
			}
			return nil
		})
	if err != nil {
		return knowledgeapi.Reconciled{}, mapError(err)
	}
	return result, nil
}

var _ knowledgeapi.Reconciler = (*Application)(nil)
