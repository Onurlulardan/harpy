package billingv1adoptionapp

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

func Compile(snapshot AdoptionSnapshot, mapping AdoptionMapping) (Bundle, error) {
	if mapping.ContractVersion != MappingContractVersion || mapping.SourceTimezone != "UTC" ||
		snapshot.Global.EnvironmentID == uuid.Nil || !canonicalInstant(snapshot.Global.SnapshotAt) {
		return Bundle{}, fmt.Errorf("Billing V1 adoption scope or mapping is invalid")
	}
	global, index, err := compileGlobal(snapshot.Global, mapping)
	if err != nil {
		return Bundle{}, err
	}
	tenants := append([]TenantSnapshot(nil), snapshot.Tenants...)
	sort.Slice(tenants, func(left, right int) bool {
		return tenants[left].TenantID.String() < tenants[right].TenantID.String()
	})
	plans := make([]ScopePlan, 0, len(tenants))
	seen := map[string]struct{}{}
	for _, tenant := range tenants {
		key := tenant.TenantID.String()
		if _, exists := seen[key]; exists {
			return Bundle{}, fmt.Errorf("duplicate Billing V1 tenant snapshot %s", key)
		}
		seen[key] = struct{}{}
		tenantMapping, exists := mapping.tenant(tenant.TenantID)
		if !exists {
			tenantMapping = TenantMapping{TenantID: tenant.TenantID}
		}
		plan, compileErr := compileTenant(tenant, tenantMapping, index, mapping.SourceTimezone)
		if compileErr != nil {
			return Bundle{}, compileErr
		}
		plans = append(plans, plan)
	}
	validateCouponCounts(&global, plans)
	finalizePlan(&global)
	return Bundle{Global: global, Tenants: plans}, nil
}
