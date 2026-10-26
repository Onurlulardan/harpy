package v1adoption

import "github.com/google/uuid"

func validateInventory(value Inventory) []Issue {
	if value.EnvironmentID == uuid.Nil ||
		(value.Mode != ModeOnPremise && value.Mode != ModeSaaS) {
		return []Issue{{Code: IssueInventoryInvalid}}
	}
	seen := map[uuid.UUID]bool{}
	issues := make([]Issue, 0)
	for _, tenant := range value.Tenants {
		if tenant.TenantID == uuid.Nil || seen[tenant.TenantID] ||
			(tenant.Status != "ACTIVE" && tenant.Status != "TRIAL" && tenant.Status != "SUSPENDED") {
			issues = append(issues, Issue{
				Code: IssueInventoryInvalid, Tenant: tenant.TenantID.String(),
				Evidence: "tenant",
			})
			continue
		}
		seen[tenant.TenantID] = true
	}
	for _, account := range value.Protected {
		if account.TenantID == uuid.Nil || !seen[account.TenantID] {
			issues = append(issues, Issue{
				Code: IssueInventoryInvalid, Tenant: account.TenantID.String(),
				Evidence: "protected_account",
			})
		}
	}
	for _, billing := range value.Billing {
		if billing.TenantID == uuid.Nil || !seen[billing.TenantID] {
			issues = append(issues, Issue{
				Code: IssueInventoryInvalid, Tenant: billing.TenantID.String(),
				Evidence: "billing",
			})
		}
	}
	return issues
}

func deploymentBoundary(value Inventory) (uuid.UUID, []TenantBinding, []Issue) {
	if value.Mode == ModeOnPremise {
		if len(value.Tenants) != 1 {
			return uuid.Nil, nil, []Issue{{Code: IssueOnPremTenantAmbiguous}}
		}
		tenant := value.Tenants[0].TenantID
		if value.Tenants[0].ControlCandidate {
			return uuid.Nil, nil, []Issue{{
				Code: IssueControlModeMismatch, Tenant: tenant.String(),
			}}
		}
		return tenant, []TenantBinding{{TenantID: tenant, Purpose: "onpremise_primary"}}, nil
	}
	var control uuid.UUID
	count := 0
	for _, tenant := range value.Tenants {
		if tenant.ControlCandidate {
			control, count = tenant.TenantID, count+1
		}
	}
	if count != 1 {
		return uuid.Nil, nil, []Issue{{Code: IssueControlTenantAmbiguous}}
	}
	bindings := []TenantBinding{{TenantID: control, Purpose: "saas_control"}}
	for _, tenant := range value.Tenants {
		if tenant.TenantID != control {
			bindings = append(bindings, TenantBinding{
				TenantID: tenant.TenantID, Purpose: "saas_customer",
			})
		}
	}
	return control, bindings, nil
}

func protectedOwner(value Inventory, primary uuid.UUID) (
	ProtectedAccountEvidence, []Issue,
) {
	candidates := make([]ProtectedAccountEvidence, 0, 1)
	for _, account := range value.Protected {
		if account.TenantID == primary && account.Active {
			candidates = append(candidates, account)
		}
	}
	if primary == uuid.Nil || len(candidates) != 1 {
		return ProtectedAccountEvidence{}, []Issue{{Code: IssueProtectedOwnerAmbiguous}}
	}
	owner := candidates[0]
	issues := make([]Issue, 0, 2)
	if !owner.CanonicalIAMMapped || !owner.ProtectedAdminBound || !validSHA(owner.MappingSHA256) {
		issues = append(issues, Issue{Code: IssueOwnerMappingMissing, Tenant: primary.String()})
	}
	if owner.UserID == uuid.Nil || !owner.UsableCredential || owner.AuthorizationVersion < 1 {
		issues = append(issues, Issue{Code: IssueOwnerCredentialMissing, Tenant: primary.String()})
	}
	return owner, issues
}
