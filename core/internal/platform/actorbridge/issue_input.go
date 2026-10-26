package actorbridge

import actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"

type IssueInput struct {
	TargetDomain         actorassertionv1.Domain
	ResourceKey          string
	ActionKey            string
	TargetBinding        [32]byte
	TokenVersion         int64
	AuthorizationVersion int64
}

func (input IssueInput) valid() bool {
	return input.TargetDomain.Validate() == nil && input.ResourceKey != "" &&
		input.ActionKey != "" && input.TargetBinding != [32]byte{} &&
		input.TokenVersion > 0 && input.AuthorizationVersion > 0
}

func (input IssueInput) decision(
	authorization Authorization,
	authority Authority,
) (actorassertionv1.DecisionEvidence, error) {
	if !input.valid() || authorization.Validate() != nil || authority.Validate() != nil ||
		authorization.TokenVersion() != input.TokenVersion ||
		authorization.AuthorizationVersion() != input.AuthorizationVersion {
		return actorassertionv1.DecisionEvidence{}, NewFault(FaultUnauthorized)
	}
	decision := actorassertionv1.DecisionEvidence{
		TenantID: authorization.TenantID(), SubjectID: authorization.SubjectID(),
		PermissionID: authorization.PermissionID(), TargetDomain: input.TargetDomain,
		ResourceKey: input.ResourceKey, ActionKey: input.ActionKey,
		TargetBindingSHA256:  input.TargetBinding,
		IAMGroupGeneration:   authority.IAMGroupGeneration(),
		IAMGeneration:        authority.IAMGeneration(),
		TargetGeneration:     authority.TargetGeneration(),
		TokenVersion:         authorization.TokenVersion(),
		AuthorizationVersion: authorization.AuthorizationVersion(),
		AuthorizationClock:   authorization.AuthorizationClock(),
	}
	if _, err := decision.Digest(); err != nil {
		return actorassertionv1.DecisionEvidence{}, NewFault(FaultInvalid)
	}
	return decision, nil
}
