package knowledgefile

import (
	"context"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Release(ctx context.Context,
	scope platformtx.TenantScope, request knowledgeapi.Release,
) error {
	mutation := ReleaseMutation{ItemID: domain.ID(request.ItemID),
		ReferenceID: domain.ID(request.ReferenceID), SourceID: domain.ID(request.SourceID),
		IdempotencyKey: request.IdempotencyKey}
	if application == nil || !validReleaseScope(ctx, scope) || !validRelease(mutation) {
		return knowledgeapi.ErrInvalid
	}
	leases, identity, err := application.leases.Bind(ctx, scope)
	if err != nil || identity.TenantID != domain.ID(scope.TenantID()) ||
		identity.ActorID != domain.ID(scope.ActorID()) {
		return knowledgeapi.ErrUnavailable
	}
	lease, replay, err := leases.PrepareRelease(ctx, mutation)
	if err != nil {
		return mapError(err)
	}
	links, linkIdentity, err := application.links.Bind(ctx, scope)
	if err != nil || linkIdentity != identity {
		return knowledgeapi.ErrUnavailable
	}
	if err = application.releaseLink(ctx, links, identity, mutation, lease, replay); err != nil {
		return mapError(err)
	}
	_, err = leases.CompleteRelease(ctx, mutation, application.clock.Now().UTC())
	return mapError(err)
}

func (application *Application) releaseLink(ctx context.Context, store linkapp.Store,
	identity foundation.Identity, mutation ReleaseMutation, lease Lease, replay bool,
) error {
	ids, err := application.nextIDs(2)
	if err != nil {
		return err
	}
	source := linkapp.Source{Owner: knowledgeapi.SourceOwner,
		Kind: knowledgeapi.SourceKind, ID: mutation.SourceID, VersionID: mutation.SourceID}
	digest := foundation.Digest([]byte(linkapp.OperationRelease), mutation.ReferenceID[:],
		mutation.SourceID[:], mutation.SourceID[:])
	result, storedReplay, err := store.Release(ctx, linkapp.Mutation{Link: domain.Link{
		ID: mutation.ReferenceID, TenantID: identity.TenantID}, ActorID: identity.ActorID,
		OperationID: ids[0], AuditID: ids[1], Source: source,
		IdempotencyKey: mutation.IdempotencyKey, RequestSHA256: digest,
		OccurredAt: application.clock.Now().UTC()})
	if err != nil || result.ID != mutation.ReferenceID || result.ItemID != lease.ItemID ||
		result.SourceID != mutation.SourceID || result.State != domain.LinkReleased ||
		replay && !storedReplay {
		return errUnavailable
	}
	return nil
}

func validReleaseScope(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		scope.OperationID() == OperationKnowledgeDelete &&
		platformtx.RequireOwner(scope, "documents") == nil
}

func validRelease(value ReleaseMutation) bool {
	return value.ItemID.UUIDv7() && value.ReferenceID.UUIDv7() &&
		value.SourceID.UUIDv7() && foundation.ValidIdempotency(value.IdempotencyKey)
}
