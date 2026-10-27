package transaction

import "github.com/google/uuid"

func AdmitCommandScope(
	actorID uuid.UUID,
	tenantID uuid.UUID,
	operationID string,
	groupID string,
	groupGeneration int64,
	domainGenerations map[string]int64,
) (ScopeRequest, error) {
	actor, err := NewActorIssuer().Issue(actorID, tenantID)
	if err != nil {
		return ScopeRequest{}, err
	}
	operation, err := NewAuthorityIssuer().Issue(
		operationID, groupID, groupGeneration, domainGenerations,
	)
	if err != nil {
		return ScopeRequest{}, err
	}
	return ScopeRequest{Actor: actor, Operation: operation, Mode: Command}, nil
}

func AdmitMultiGroupCommandScope(
	actorID uuid.UUID,
	tenantID uuid.UUID,
	operationID string,
	groups []CommandAuthorityGroup,
	domains []CommandAuthorityDomain,
	participants []string,
) (ScopeRequest, error) {
	actor, err := NewActorIssuer().Issue(actorID, tenantID)
	if err != nil {
		return ScopeRequest{}, err
	}
	operation, err := NewAuthorityIssuer().IssueMultiGroup(
		operationID, groups, domains, participants,
	)
	if err != nil {
		return ScopeRequest{}, err
	}
	return ScopeRequest{Actor: actor, Operation: operation, Mode: Command}, nil
}
