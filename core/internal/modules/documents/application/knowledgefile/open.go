package knowledgefile

import (
	"context"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Open(ctx context.Context,
	request knowledgeapi.OpenRequest,
) (knowledgeapi.Opened, error) {
	access := Access{TenantID: domain.ID(request.TenantID),
		ActorID:              domain.ID(request.ActorID),
		AuthorizationVersion: request.AuthorizationVersion}
	itemID, versionID := domain.ID(request.ItemID), domain.ID(request.VersionID)
	if application == nil || !access.TenantID.Valid() || !access.ActorID.Valid() ||
		!itemID.UUIDv7() || !versionID.UUIDv7() || !platformtx.NetworkAllowed(ctx) {
		return knowledgeapi.Opened{}, knowledgeapi.ErrInvalid
	}
	err := application.scopes.InSnapshot(ctx, access, downloadapp.OperationVersion,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			store, identity, bindErr := application.leases.Bind(scoped, scope)
			if bindErr != nil || identity.TenantID != access.TenantID ||
				identity.ActorID != access.ActorID {
				return errUnavailable
			}
			_, bindErr = store.VerifyOpen(scoped, itemID, versionID,
				application.clock.Now().UTC())
			return bindErr
		})
	if err != nil {
		return knowledgeapi.Opened{}, mapError(err)
	}
	result, err := application.downloads.Open(ctx, boundRunner{application, access},
		downloadapp.Request{ItemID: itemID, VersionID: versionID})
	if err != nil {
		return knowledgeapi.Opened{}, mapError(err)
	}
	if result.ItemID != itemID || result.VersionID != versionID ||
		result.Content == nil || result.Bytes < 1 || result.SHA256 == [32]byte{} {
		return knowledgeapi.Opened{}, knowledgeapi.ErrUnavailable
	}
	return knowledgeapi.Opened{Content: result.Content,
		Bytes: result.Bytes, SHA256: result.SHA256}, nil
}
