package dashboard

import (
	"context"
	"encoding/binary"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type PublishRequest struct {
	DashboardID, DraftVersionID                 uuid.UUID
	ExpectedRootRevision, ExpectedDraftRevision uint64
	IdempotencyKey                              string
}

func (application *Application) Publish(
	ctx context.Context, scope platformtx.TenantScope, request PublishRequest,
) (Snapshot, error) {
	if !validUUIDv7(request.DashboardID) || !validUUIDv7(request.DraftVersionID) ||
		request.ExpectedRootRevision == 0 || request.ExpectedDraftRevision == 0 ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationPublish, PermissionPublish)
	if err != nil {
		return Snapshot{}, err
	}
	return application.publishBound(ctx, scope, store, identity, request)
}

func (application *Application) publishBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request PublishRequest,
) (Snapshot, error) {
	current, err := store.Get(ctx, request.DashboardID)
	if err != nil || current.VersionID != request.DraftVersionID || current.Published ||
		validateSnapshot(current) != nil || len(current.Widgets) == 0 {
		return Snapshot{}, ErrConflict
	}
	queryVersions := make([]uuid.UUID, len(current.Widgets))
	for index, widget := range current.Widgets {
		queryVersions[index] = widget.QueryVersionID
	}
	manifest, err := application.queries.VerifyPublished(ctx, scope, queryVersions)
	if err != nil || manifest == [32]byte{} {
		return Snapshot{}, ErrConflict
	}
	ids, err := application.nextIDs(2)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	revisions := make([]byte, 16)
	binary.BigEndian.PutUint64(revisions[:8], request.ExpectedRootRevision)
	binary.BigEndian.PutUint64(revisions[8:], request.ExpectedDraftRevision)
	requestDigest := digest([]byte(OperationPublish), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.DashboardID[:], request.DraftVersionID[:],
		revisions, manifest[:])
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourcePublish, ids[0], request.DashboardID,
		request.IdempotencyKey, requestDigest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Publish(ctx, PublishMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, DashboardID: decision.OwnerFactID,
		DraftVersionID: request.DraftVersionID, NextDraftVersionID: ids[1],
		IdempotencyKey: request.IdempotencyKey,
		RequestSHA256:  requestDigest, ExpectedRootRevision: request.ExpectedRootRevision,
		ExpectedDraftRevision: request.ExpectedDraftRevision, ManifestSHA256: manifest,
		OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("publish analytics dashboard: %w", err)
	}
	if validateSnapshot(result) != nil || result.DashboardID != decision.OwnerFactID ||
		!result.Published || result.ManifestSHA256 != manifest || replay != decision.Replay ||
		(!replay && result.VersionID != request.DraftVersionID) {
		return Snapshot{}, ErrUnavailable
	}
	if !replay && application.effects.DashboardPublished(ctx, scope, PublishEffect{
		OperationID: decision.OperationID, ActorID: identity.ActorID, TenantID: identity.TenantID,
		DashboardID: result.DashboardID, DashboardVersionID: result.VersionID,
		RootRevision: result.RootRevision, VersionNumber: result.VersionNumber,
		ManifestSHA256: result.ManifestSHA256,
		OccurredAt:     now,
	}) != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
