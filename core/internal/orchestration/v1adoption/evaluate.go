package v1adoption

import (
	"sort"

	"github.com/google/uuid"
)

func Evaluate(inventory Inventory) (Plan, error) {
	canonical := canonicalInventory(inventory)
	issues := validateInventory(canonical)
	primary, bindings, boundaryIssues := deploymentBoundary(canonical)
	issues = append(issues, boundaryIssues...)
	owner, ownerIssues := protectedOwner(canonical, primary)
	issues = append(issues, ownerIssues...)
	assignments, billingIssues := baseAssignments(canonical, bindings)
	issues = append(issues, billingIssues...)
	sortIssues(issues)
	if len(issues) != 0 {
		return Plan{}, &FenceError{Issues: issues}
	}
	inventorySHA := digest(canonical)
	plan := Plan{
		EnvironmentID: inventory.EnvironmentID, Mode: inventory.Mode,
		PrimaryTenantID: primary, OwnerUserID: owner.UserID,
		OwnerAuthorizationVersion: owner.AuthorizationVersion,
		OwnerMappingSHA:           owner.MappingSHA256, TenantBindings: bindings,
		BaseAssignments: assignments, InventorySHA256: inventorySHA,
	}
	plan.PlanSHA256 = digest(struct {
		EnvironmentID             uuid.UUID
		Mode                      DeploymentMode
		PrimaryTenantID           uuid.UUID
		OwnerUserID               uuid.UUID
		OwnerAuthorizationVersion int64
		OwnerMappingSHA           string
		TenantBindings            []TenantBinding
		BaseAssignments           []BaseAssignment
	}{plan.EnvironmentID, plan.Mode, plan.PrimaryTenantID, plan.OwnerUserID,
		plan.OwnerAuthorizationVersion,
		plan.OwnerMappingSHA, plan.TenantBindings, plan.BaseAssignments})
	return plan, nil
}

func canonicalInventory(value Inventory) Inventory {
	copyValue := value
	copyValue.Tenants = append([]TenantEvidence(nil), value.Tenants...)
	copyValue.Protected = append([]ProtectedAccountEvidence(nil), value.Protected...)
	copyValue.Billing = append([]BillingEvidence(nil), value.Billing...)
	sort.Slice(copyValue.Tenants, func(i, j int) bool {
		left, right := copyValue.Tenants[i], copyValue.Tenants[j]
		if left.TenantID != right.TenantID {
			return left.TenantID.String() < right.TenantID.String()
		}
		return digest(left) < digest(right)
	})
	sort.Slice(copyValue.Protected, func(i, j int) bool {
		left, right := copyValue.Protected[i], copyValue.Protected[j]
		if left.TenantID != right.TenantID {
			return left.TenantID.String() < right.TenantID.String()
		}
		if left.UserID != right.UserID {
			return left.UserID.String() < right.UserID.String()
		}
		return digest(left) < digest(right)
	})
	sort.Slice(copyValue.Billing, func(i, j int) bool {
		left, right := copyValue.Billing[i], copyValue.Billing[j]
		if left.TenantID != right.TenantID {
			return left.TenantID.String() < right.TenantID.String()
		}
		return digest(left) < digest(right)
	})
	return copyValue
}
