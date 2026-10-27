package billingv1adoptionapp

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

func convertGrants(
	limits map[string]int64,
	features map[string]bool,
) ([]offeringdomain.Grant, error) {
	values := map[entitlement.Key]entitlement.Value{}
	var err error
	values["iam.active_users"], err = integerValue(limits["max_users"])
	if err != nil {
		return nil, ErrLimits
	}
	values["documents.storage_bytes"], err = storageValue(limits["max_storage_mb"])
	if err != nil {
		return nil, ErrLimits
	}
	values["workflow.enabled"] = entitlement.BooleanValue(features["workflows"])
	values["api.access"] = entitlement.BooleanValue(features["api_access"])
	registry := entitlement.RegistryV1()
	grants := make([]offeringdomain.Grant, 0, len(registry))
	for _, entry := range registry {
		grant, grantErr := offeringdomain.NewGrant(
			entry.Definition, values[entry.Definition.Key()],
		)
		if grantErr != nil {
			return nil, ErrPlan
		}
		grants = append(grants, grant)
	}
	return grants, nil
}
