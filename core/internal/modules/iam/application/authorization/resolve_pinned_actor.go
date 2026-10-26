package authorization

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authorizer Authorizer) ResolvePinnedActor(
	ctx context.Context,
	scope transaction.TenantScope,
	request PinnedActorRequest,
) (PinnedActorDecision, error) {
	invalid := newPinnedActorDecision(
		deniedDecision(ReasonInvalidRequest, 0, 0), 0,
	)
	if !authorizer.valid || evidencePortIsNil(authorizer.port) ||
		request.Validate() != nil || !validAuthorizationScope(ctx, scope) {
		return invalid, ErrAuthorizationContract
	}
	probe, err := NewRequest(request.ResourceKey(), request.ActionKey(),
		1, request.AuthorizationVersion())
	if err != nil {
		return invalid, ErrAuthorizationContract
	}
	snapshot, err := authorizer.port.Authorize(ctx, scope, probe)
	if err != nil || snapshot.Validate() != nil {
		unavailable := newPinnedActorDecision(
			deniedDecision(ReasonEvidenceUnavailable, 0, 0), 0,
		)
		return unavailable, fmt.Errorf("%w", ErrAuthorizationEvidence)
	}
	return evaluatePinnedActor(request, snapshot), nil
}

func evaluatePinnedActor(
	request PinnedActorRequest,
	snapshot Snapshot,
) PinnedActorDecision {
	version, token := int64(0), int64(0)
	if snapshot.actor != nil {
		version, token = snapshot.actor.AuthorizationVersion, snapshot.actor.TokenVersion
	}
	denied := func(reason Reason) PinnedActorDecision {
		return newPinnedActorDecision(deniedDecision(reason, snapshot.clock, version), token)
	}
	if snapshot.actor == nil || snapshot.actor.Status != "active" {
		return denied(ReasonActorUnavailable)
	}
	if version != request.AuthorizationVersion() {
		return denied(ReasonAuthorizationStale)
	}
	if snapshot.resource == nil || snapshot.resource.State != "active" {
		return denied(ReasonResourceUnavailable)
	}
	if snapshot.action == nil || snapshot.action.State != "active" {
		return denied(ReasonActionUnavailable)
	}
	reason, permissionID, allowed := effectiveGrant(snapshot)
	if !allowed {
		return denied(ReasonNoEffectivePermission)
	}
	return newPinnedActorDecision(
		allowedDecision(reason, snapshot.clock, version, permissionID), token,
	)
}
