package transaction

import (
	"fmt"
	"slices"
)

const maximumSnapshotOwners = 4

type SnapshotAuthorityIssuer struct {
	valid bool
}

func NewSnapshotAuthorityIssuer() SnapshotAuthorityIssuer {
	return SnapshotAuthorityIssuer{valid: true}
}

func (issuer SnapshotAuthorityIssuer) Issue(
	operationID string,
	domainGenerations map[string]int64,
) (VerifiedOperation, error) {
	if !issuer.valid || !operationIDPattern.MatchString(operationID) ||
		len(domainGenerations) < 1 || len(domainGenerations) > maximumSnapshotOwners {
		return VerifiedOperation{}, fmt.Errorf("snapshot operation capability is invalid")
	}
	names := make([]string, 0, len(domainGenerations))
	for name, generation := range domainGenerations {
		if !authorityNamePattern.MatchString(name) || generation <= 0 {
			return VerifiedOperation{}, fmt.Errorf("snapshot operation capability is invalid")
		}
		names = append(names, name)
	}
	slices.Sort(names)
	domains := make([]domainExpectation, 0, len(names))
	for _, name := range names {
		domains = append(domains, domainExpectation{
			name: name, generation: domainGenerations[name],
		})
	}
	return VerifiedOperation{
		operationID: operationID, domains: domains,
		mode: SnapshotQuery, issued: true,
	}, nil
}

func validSnapshotOperation(operation VerifiedOperation) bool {
	maximum := maximumSnapshotOwners
	if tenantSettingsSnapshotOperation(operation.operationID) {
		maximum = tenantSettingsSnapshotOwnerCount
	}
	if len(operation.groups) != 0 || len(operation.participants) != 0 ||
		len(operation.domains) < 1 || len(operation.domains) > maximum {
		return false
	}
	previous := ""
	for _, domain := range operation.domains {
		if !authorityNamePattern.MatchString(domain.name) || domain.generation <= 0 ||
			domain.groupID != "" ||
			previous >= domain.name {
			return false
		}
		previous = domain.name
	}
	return true
}
