package cutoverv1

import "slices"

var tenantGroupDomains = map[string][]string{
	"iam_system_resources": {"iam", "system_resource_layout"},
	"billing_forms":        {"billing", "forms"},
	"workflow":             {"workflow"},
	"analytics":            {"analytics"},
	"catalog":              {"catalog"},
	"work":                 {"work"},
	"documents":            {"documents"},
	"notifications":        {"notifications"},
}

func validateLeaseScope(lease CredentialLease) error {
	if !sqlNamePattern.MatchString(lease.GroupID) || len(lease.Domains) == 0 ||
		len(lease.Domains) > 2 {
		return contractError("cutover scope")
	}
	if lease.ScopeKind == "global" {
		if lease.GroupID != "billing" || !slices.Equal(lease.Domains, []string{"billing"}) {
			return contractError("global billing scope")
		}
		return nil
	}
	expected, exists := tenantGroupDomains[lease.GroupID]
	if lease.ScopeKind != "tenant" || !exists || !slices.Equal(lease.Domains, expected) {
		return contractError("tenant cutover group")
	}
	return nil
}
