package iam

import "fmt"

type CustomerProvisioningAccessRuntime struct{ base *BillingAccessRuntime }

func NewCustomerProvisioningAccessRuntime(
	base *BillingAccessRuntime,
) (*CustomerProvisioningAccessRuntime, error) {
	if base == nil || base.pool == nil || base.coordinator == nil ||
		!base.security.Valid() || base.platformIAM == nil || base.epoch == nil ||
		base.clock == nil {
		return nil, fmt.Errorf("IAM customer provisioning access dependencies are required")
	}
	return &CustomerProvisioningAccessRuntime{base: base}, nil
}
