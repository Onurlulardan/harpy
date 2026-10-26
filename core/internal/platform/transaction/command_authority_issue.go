package transaction

import "fmt"

type CommandAuthorityGroup struct {
	ID         string
	Generation int64
}

type CommandAuthorityDomain struct {
	Name       string
	Generation int64
}

func (issuer AuthorityIssuer) Issue(
	operationID string,
	groupID string,
	groupGeneration int64,
	domainGenerations map[string]int64,
) (VerifiedOperation, error) {
	domains := make([]CommandAuthorityDomain, 0, len(domainGenerations))
	for name, generation := range domainGenerations {
		domains = append(domains, CommandAuthorityDomain{
			Name: name, Generation: generation,
		})
	}
	members, _ := tenantAuthorityGroupMembers(groupID)
	participants := append([]string(nil), members...)
	return issuer.IssueMultiGroup(
		operationID,
		[]CommandAuthorityGroup{{ID: groupID, Generation: groupGeneration}},
		domains,
		participants,
	)
}

func (issuer AuthorityIssuer) IssueMultiGroup(
	operationID string,
	groups []CommandAuthorityGroup,
	domains []CommandAuthorityDomain,
	participants []string,
) (VerifiedOperation, error) {
	if !issuer.valid || !operationIDPattern.MatchString(operationID) {
		return VerifiedOperation{}, invalidCommandAuthority()
	}
	canonicalGroups, owners, err := canonicalCommandGroups(groups)
	if err != nil {
		return VerifiedOperation{}, err
	}
	canonicalDomains, err := canonicalCommandDomains(domains, owners)
	if err != nil {
		return VerifiedOperation{}, err
	}
	canonicalParticipants, err := canonicalParticipants(participants, owners)
	if err != nil {
		return VerifiedOperation{}, err
	}
	operation := VerifiedOperation{
		operationID: operationID, groups: canonicalGroups,
		domains: canonicalDomains, participants: canonicalParticipants,
		mode: Command, issued: true,
	}
	if !operation.valid() {
		return VerifiedOperation{}, invalidCommandAuthority()
	}
	return operation, nil
}

func invalidCommandAuthority() error {
	return fmt.Errorf("verified operation capability is invalid")
}
