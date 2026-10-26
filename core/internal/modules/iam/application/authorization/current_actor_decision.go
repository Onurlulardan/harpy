package authorization

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authorizer Authorizer) CurrentActorDecision(
	ctx context.Context,
	scope transaction.TenantScope,
	request CurrentActorRequest,
) (Decision, error) {
	invalid := deniedDecision(ReasonInvalidRequest, 0, 0)
	if !authorizer.valid || evidencePortIsNil(authorizer.port) ||
		request.Validate() != nil || !validAuthorizationScope(ctx, scope) {
		return invalid, ErrAuthorizationContract
	}
	probe, err := newCurrentActorEvidenceRequest(
		request.ResourceKey(), request.ActionKey(), request.ActorID(),
	)
	if err != nil {
		return invalid, ErrAuthorizationContract
	}
	snapshot, err := authorizer.port.Authorize(ctx, scope, probe)
	if err != nil || snapshot.Validate() != nil {
		return deniedDecision(ReasonEvidenceUnavailable, 0, 0),
			fmt.Errorf("%w", ErrAuthorizationEvidence)
	}
	return evaluateCurrentActor(snapshot), nil
}

func evaluateCurrentActor(snapshot Snapshot) Decision {
	version := int64(0)
	if snapshot.actor != nil {
		version = snapshot.actor.AuthorizationVersion
	}
	denied := func(reason Reason) Decision {
		return deniedDecision(reason, snapshot.clock, version)
	}
	if snapshot.actor == nil || snapshot.actor.Status != "active" {
		return denied(ReasonActorUnavailable)
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
	return allowedDecision(reason, snapshot.clock, version, permissionID)
}
