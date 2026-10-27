package knowledgefile

import (
	"context"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Adopt(ctx context.Context,
	scope platformtx.TenantScope, request knowledgeapi.Adoption,
) error {
	mutation := AdoptMutation{LeaseID: domain.ID(request.LeaseID),
		ItemID: domain.ID(request.ItemID), VersionID: domain.ID(request.VersionID),
		ReferenceID: domain.ID(request.ReferenceID), SourceID: domain.ID(request.SourceID),
		IdempotencyKey: request.IdempotencyKey}
	if application == nil || !validAdoptScope(ctx, scope) || !validAdopt(mutation) {
		return knowledgeapi.ErrInvalid
	}
	leases, identity, err := application.leases.Bind(ctx, scope)
	if err != nil || identity.TenantID != domain.ID(scope.TenantID()) ||
		identity.ActorID != domain.ID(scope.ActorID()) {
		return knowledgeapi.ErrUnavailable
	}
	lease, replay, err := leases.PrepareAdopt(ctx, mutation, application.clock.Now().UTC())
	if err != nil {
		return mapError(err)
	}
	links, linkIdentity, err := application.links.Bind(ctx, scope)
	if err != nil || linkIdentity != identity {
		return knowledgeapi.ErrUnavailable
	}
	if err = application.registerLink(ctx, links, identity, mutation, lease, replay); err != nil {
		return mapError(err)
	}
	_, err = leases.CompleteAdopt(ctx, mutation, application.clock.Now().UTC())
	return mapError(err)
}

func (application *Application) registerLink(ctx context.Context, store linkapp.Store,
	identity foundation.Identity, mutation AdoptMutation, lease Lease, replay bool,
) error {
	epoch, err := store.LockActiveEpoch(ctx, mutation.ItemID)
	if err != nil {
		return err
	}
	value, err := domain.NewLink(mutation.ReferenceID, identity.TenantID,
		mutation.ItemID, mutation.SourceID, mutation.SourceID,
		knowledgeapi.SourceOwner, knowledgeapi.SourceKind, knowledgeapi.Usage,
		epoch, application.clock.Now().UTC())
	if err != nil {
		return knowledgeapi.ErrInvalid
	}
	ids, err := application.nextIDs(2)
	if err != nil {
		return err
	}
	source := linkapp.Source{Owner: knowledgeapi.SourceOwner,
		Kind: knowledgeapi.SourceKind, ID: mutation.SourceID, VersionID: mutation.SourceID}
	digest := foundation.Digest([]byte(linkapp.OperationRegister), mutation.ItemID[:],
		mutation.SourceID[:], mutation.SourceID[:], []byte(source.Owner),
		[]byte(source.Kind), []byte(knowledgeapi.Usage))
	result, storedReplay, err := store.Register(ctx, linkapp.Mutation{Link: value,
		ActorID: identity.ActorID, OperationID: ids[0], AuditID: ids[1],
		Source: source, IdempotencyKey: mutation.IdempotencyKey,
		RequestSHA256: digest, OccurredAt: application.clock.Now().UTC()})
	if err != nil || result.ID != mutation.ReferenceID || result.ItemID != lease.ItemID ||
		result.SourceID != mutation.SourceID || result.State != domain.LinkActive ||
		replay && !storedReplay {
		return errUnavailable
	}
	return nil
}

func validAdoptScope(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		scope.OperationID() == OperationKnowledgeCreate &&
		platformtx.RequireOwner(scope, "documents") == nil
}

func validAdopt(value AdoptMutation) bool {
	return value.LeaseID.UUIDv7() && value.ItemID.UUIDv7() && value.VersionID.UUIDv7() &&
		value.ReferenceID.UUIDv7() && value.SourceID.UUIDv7() &&
		foundation.ValidIdempotency(value.IdempotencyKey)
}
