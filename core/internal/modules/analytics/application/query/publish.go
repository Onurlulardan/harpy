package query

import (
	"context"
	"fmt"

	analyticsforms "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/formsreference"
	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Publish(
	ctx context.Context, scope platformtx.TenantScope, request PublishRequest,
) (Snapshot, error) {
	if !validPublishRequest(request) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationPublish, PermissionPublish)
	if err != nil {
		return Snapshot{}, err
	}
	return application.publishBound(ctx, scope, store, identity, OperationPublish, request)
}

func (application *Application) publishBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, operation string, request PublishRequest,
) (Snapshot, error) {
	current, err := store.Get(ctx, request.QueryID)
	if err != nil || current.VersionID != request.DraftVersionID ||
		current.VersionState != analyticsquery.VersionDraft || validateSnapshot(current) != nil {
		return Snapshot{}, ErrConflict
	}
	contract, dependencies, manifest, err := application.resolvePublication(
		ctx, scope, current, request)
	if err != nil {
		return Snapshot{}, err
	}
	operationID, nextDraftID, err := application.twoIDs()
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	digest := publishDigest(operation, identity, request, contract, manifest)
	cachePolicy := analyticspublication.CacheNone
	if request.RevisionBoundCache {
		cachePolicy = analyticspublication.CacheRevisionBound
	}
	mutation := billingapi.MeteredResourcePublish
	if operation == OperationRevalidate {
		mutation = billingapi.MeteredResourceRevalidate
	} else if operation != OperationPublish {
		return Snapshot{}, ErrUnavailable
	}
	decision, err := application.decideMutation(ctx, scope, identity, mutation,
		operationID, request.QueryID, request.IdempotencyKey, digest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Publish(ctx, PublishMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, QueryID: decision.OwnerFactID,
		DraftVersionID: request.DraftVersionID, NextDraftVersionID: nextDraftID,
		IdempotencyKey: request.IdempotencyKey,
		RequestSHA256:  digest, ExpectedRootRevision: request.ExpectedRootRevision,
		ExpectedDraftRevision: request.ExpectedDraftRevision, Contract: contract,
		Dependencies: dependencies, ManifestSHA256: manifest,
		CachePolicy: cachePolicy, OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("publish analytics query: %w", err)
	}
	if validateSnapshot(result) != nil || result.QueryID != decision.OwnerFactID ||
		result.VersionState != analyticsquery.VersionPublished ||
		result.ManifestSHA256 != manifest || replay != decision.Replay ||
		(!replay && result.VersionID != request.DraftVersionID) {
		return Snapshot{}, ErrUnavailable
	}
	targets, err := formsReferenceTargets(dependencies)
	if err != nil {
		return Snapshot{}, err
	}
	if err = application.references.BindVersion(ctx, scope, analyticsforms.Source{
		Kind: "query", ID: analyticsforms.ID(result.QueryID),
		VersionID: analyticsforms.ID(result.VersionID),
	}, targets, now); err != nil {
		return Snapshot{}, fmt.Errorf("bind Analytics Forms references: %w", err)
	}
	if !replay && application.effects.QueryPublished(ctx, scope, PublishEffect{
		OperationID: decision.OperationID, ActorID: identity.ActorID, TenantID: identity.TenantID,
		QueryID: result.QueryID, QueryVersionID: result.VersionID,
		RootRevision: result.RootRevision, VersionNumber: result.VersionNumber,
		ManifestSHA256: result.ManifestSHA256,
		OccurredAt:     now, Revalidated: operation == OperationRevalidate,
	}) != nil {
		return Snapshot{}, ErrUnavailable
	}
	application.invalidate(ctx, request.QueryID)
	return result, nil
}

func formsReferenceTargets(
	dependencies []DependencySnapshot,
) ([]analyticsforms.Target, error) {
	targets := make([]analyticsforms.Target, 0, len(dependencies))
	for _, dependency := range dependencies {
		if dependency.OwnerDomain != "forms" {
			continue
		}
		if dependency.ResourceKind != "form" {
			return nil, ErrUnavailable
		}
		targets = append(targets, analyticsforms.Target{
			DefinitionID: analyticsforms.ID(dependency.ResourceID),
			VersionID:    analyticsforms.ID(dependency.VersionID),
		})
	}
	return targets, nil
}

func (application *Application) twoIDs() (first, second uuid.UUID, err error) {
	first, err = application.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(first) {
		return uuid.Nil, uuid.Nil, ErrUnavailable
	}
	second, err = application.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(second) || first == second {
		return uuid.Nil, uuid.Nil, ErrUnavailable
	}
	return first, second, nil
}
