package publication

import (
	"context"
	"crypto/sha256"
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ActivateRequest struct {
	ReleaseID        uuid.UUID
	ExpectedRevision uint64
	CapabilityTTL    time.Duration
	IdempotencyKey   string
}

func (application *Application) Activate(
	ctx context.Context, scope platformtx.TenantScope, request ActivateRequest,
) (Activation, error) {
	if !validUUIDv7(request.ReleaseID) || request.ExpectedRevision == 0 ||
		(request.CapabilityTTL < 0 || request.CapabilityTTL > 30*24*time.Hour) ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return Activation{}, ErrInvalid
	}
	store, identity, current, timing, err := application.transitionStore(ctx, scope,
		OperationActivate, request.ReleaseID, false)
	if err != nil {
		return Activation{}, err
	}
	if current.State != analyticspublication.StatePendingApproval &&
		current.State != analyticspublication.StateActive ||
		(current.Audience != analyticspublication.AudienceTenant && current.ApproverID == uuid.Nil) {
		return Activation{}, ErrConflict
	}
	return application.activateBound(ctx, scope, store, identity, current, timing, request)
}

func (application *Application) activateBound(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	identity ScopeIdentity,
	current Snapshot,
	timing timeBinding,
	request ActivateRequest,
) (Activation, error) {
	operationID, err := application.nextID()
	if err != nil {
		return Activation{}, err
	}
	requestHash := activationDigest(identity, request)
	decision, err := application.decideMutation(ctx, scope, identity,
		current.ResourceKind, operationID, request.ReleaseID,
		request.IdempotencyKey, requestHash)
	if err != nil {
		return Activation{}, err
	}
	publicID := [16]byte{}
	capabilityHash := [32]byte{}
	var capabilityToken []byte
	var capabilityExpiry *time.Time
	if !decision.Replay && current.Audience != analyticspublication.AudienceTenant {
		random, err := application.randomBytes(16)
		if err != nil {
			return Activation{}, err
		}
		copy(publicID[:], random)
	}
	if !decision.Replay && current.Audience == analyticspublication.AudienceCapability {
		if request.CapabilityTTL <= 0 {
			return Activation{}, ErrInvalid
		}
		capabilityToken, err = application.randomBytes(32)
		if err != nil {
			return Activation{}, err
		}
		capabilityHash = sha256.Sum256(capabilityToken)
		expires := timing.now.Add(request.CapabilityTTL)
		capabilityExpiry = &expires
	}
	result, replay, err := store.Activate(ctx, ActivateMutation{TransitionMutation: TransitionMutation{
		OperationID: decision.OperationID, ActorID: identity.ActorID, ReleaseID: decision.OwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestHash,
		ExpectedRevision: request.ExpectedRevision, StepUpAt: timing.stepUp,
		OccurredAt: timing.now}, PublicID: publicID, CapabilitySHA256: capabilityHash,
		CapabilityExpiresAt: capabilityExpiry})
	if err != nil {
		return Activation{}, err
	}
	if validateSnapshot(result) != nil || result.ReleaseID != decision.OwnerFactID ||
		result.State != analyticspublication.StateActive || replay != decision.Replay ||
		(!replay && (result.Revision != request.ExpectedRevision+1 || result.PublicID != publicID)) {
		return Activation{}, ErrUnavailable
	}
	if !replay && application.effects.Activated(ctx, scope, TransitionEffect{
		OperationID: decision.OperationID, ActorID: identity.ActorID, TenantID: identity.TenantID,
		ReleaseID: result.ReleaseID, ResourceVersionID: result.ResourceVersionID,
		ResourceKind: result.ResourceKind, Audience: result.Audience,
		Generation: result.Generation, OccurredAt: timing.now,
	}) != nil {
		return Activation{}, ErrUnavailable
	}
	if replay {
		capabilityToken = nil
	}
	return Activation{Snapshot: result, CapabilityToken: capabilityToken, Replay: replay}, nil
}
