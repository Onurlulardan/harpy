package transaction

import (
	"fmt"
	"slices"
)

const tenantSettingsSnapshotOwnerCount = 4

type TenantSettingsSnapshotIssuer struct {
	valid bool
}

func NewTenantSettingsSnapshotIssuer() TenantSettingsSnapshotIssuer {
	return TenantSettingsSnapshotIssuer{valid: true}
}

func tenantSettingsSnapshotOperation(operationID string) bool {
	return operationID == "tenancy.settings.get"
}

func (issuer TenantSettingsSnapshotIssuer) Issue(
	operationID string,
	domainGenerations map[string]int64,
) (VerifiedOperation, error) {
	if !issuer.valid || !tenantSettingsSnapshotOperation(operationID) ||
		len(domainGenerations) != tenantSettingsSnapshotOwnerCount {
		return VerifiedOperation{}, fmt.Errorf("tenant settings snapshot is invalid")
	}
	required := []string{"billing", "iam", "notifications", "tenancy"}
	names := make([]string, 0, len(domainGenerations))
	for name, generation := range domainGenerations {
		if !authorityNamePattern.MatchString(name) || generation <= 0 {
			return VerifiedOperation{}, fmt.Errorf("tenant settings snapshot is invalid")
		}
		names = append(names, name)
	}
	slices.Sort(names)
	if !slices.Equal(names, required) {
		return VerifiedOperation{}, fmt.Errorf("tenant settings snapshot is invalid")
	}
	domains := make([]domainExpectation, 0, len(names))
	for _, name := range names {
		domains = append(domains, domainExpectation{
			name: name, generation: domainGenerations[name],
		})
	}
	operation := VerifiedOperation{
		operationID: operationID, domains: domains,
		mode: SnapshotQuery, issued: true,
	}
	if !operation.valid() {
		return VerifiedOperation{}, fmt.Errorf("tenant settings snapshot is invalid")
	}
	return operation, nil
}
