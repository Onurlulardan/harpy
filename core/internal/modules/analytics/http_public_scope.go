package analytics

import (
	"context"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type publicScopeCoordinator struct {
	resolver    HTTPGenerationResolver
	coordinator HTTPScopeCoordinator
}

func newPublicScopeCoordinator(
	resolver HTTPGenerationResolver, coordinator HTTPScopeCoordinator,
) (*publicScopeCoordinator, error) {
	if nilHTTPPort(resolver) || nilHTTPPort(coordinator) {
		return nil, publicationapp.ErrInvalid
	}
	return &publicScopeCoordinator{resolver: resolver, coordinator: coordinator}, nil
}

func (coordinator *publicScopeCoordinator) InPublicationSnapshot(
	ctx context.Context, route publicationapp.PublicRoute,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if coordinator == nil || nilHTTPPort(action) || ctx == nil || ctx.Err() != nil ||
		route.TenantID == uuid.Nil || route.ReleaseID == uuid.Nil ||
		route.PublicID == [16]byte{} || route.Generation == 0 {
		return publicationapp.ErrUnavailable
	}
	_, analyticsGeneration, err := coordinator.resolver.ResolveTenantAnalyticsGenerations(
		ctx, route.TenantID)
	if err != nil || analyticsGeneration < 1 {
		return publicationapp.ErrUnavailable
	}
	_, _, formsGeneration, err := coordinator.resolver.ResolveTenantBillingGenerations(
		ctx, route.TenantID)
	if err != nil || formsGeneration < 1 {
		return publicationapp.ErrUnavailable
	}
	identity := append([]byte("harpy/analytics/public-principal/v1\x00"),
		route.ReleaseID[:]...)
	identity = append(identity, route.PublicID[:]...)
	principalID := uuid.NewSHA1(uuid.NameSpaceOID, identity)
	actor, err := actorcontext.Issue(principalID, route.TenantID)
	if err != nil {
		return publicationapp.ErrUnavailable
	}
	operation, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		"analytics.publication.execute", map[string]int64{
			"analytics": analyticsGeneration, "forms": formsGeneration,
		})
	if err != nil {
		return publicationapp.ErrUnavailable
	}
	return coordinator.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: operation, Mode: platformtx.SnapshotQuery}, action)
}

var _ publicationapp.PublicScopeCoordinator = (*publicScopeCoordinator)(nil)
