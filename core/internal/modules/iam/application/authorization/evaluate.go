package authorization

import "github.com/google/uuid"

func evaluate(request Request, snapshot Snapshot) Decision {
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
	if snapshot.actor.TokenVersion != request.TokenVersion() {
		return denied(ReasonTokenStale)
	}
	if snapshot.actor.AuthorizationVersion != request.AuthorizationVersion() {
		return denied(ReasonAuthorizationStale)
	}
	if snapshot.resource == nil || snapshot.resource.State != "active" {
		return denied(ReasonResourceUnavailable)
	}
	if snapshot.action == nil || snapshot.action.State != "active" {
		return denied(ReasonActionUnavailable)
	}
	if reason, permissionID, allowed := effectiveGrant(snapshot); allowed {
		return allowedDecision(reason, snapshot.clock, version, permissionID)
	}
	return denied(ReasonNoEffectivePermission)
}

func effectiveGrant(snapshot Snapshot) (Reason, uuid.UUID, bool) {
	bestRank := 0
	bestReason := Reason("")
	bestPermissionID := uuid.Nil
	for _, fact := range snapshot.bindings {
		reason, rank, allowed := eligibleBinding(
			fact, snapshot.resource.ID, snapshot.action.ID,
		)
		if allowed && (bestRank == 0 || rank < bestRank) {
			bestRank, bestReason = rank, reason
			bestPermissionID = fact.PermissionID
		}
	}
	for _, edge := range snapshot.implications {
		if !edge.activeFor(snapshot.resource.ID, snapshot.action.ID) {
			continue
		}
		for _, fact := range snapshot.bindings {
			reason, rank, allowed := eligibleImplication(fact, edge)
			if allowed && (bestRank == 0 || rank < bestRank) {
				bestRank, bestReason = rank, reason
				bestPermissionID = fact.PermissionID
			}
		}
	}
	return bestReason, bestPermissionID, bestRank > 0
}
