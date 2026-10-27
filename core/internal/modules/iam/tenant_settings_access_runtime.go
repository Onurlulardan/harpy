package iam

import "fmt"

type TenantSettingsAccessRuntime struct{ base *BillingAccessRuntime }

func NewTenantSettingsAccessRuntime(
	base *BillingAccessRuntime,
) (*TenantSettingsAccessRuntime, error) {
	if base == nil || base.pool == nil || base.coordinator == nil ||
		!base.security.Valid() || base.epoch == nil ||
		base.clock == nil {
		return nil, fmt.Errorf("IAM tenant settings access dependencies are required")
	}
	return &TenantSettingsAccessRuntime{base: base}, nil
}
