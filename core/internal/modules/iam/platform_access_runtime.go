package iam

import "fmt"

type PlatformAccessRuntime struct{ base *BillingAccessRuntime }

func NewPlatformAccessRuntime(
	base *BillingAccessRuntime,
) (*PlatformAccessRuntime, error) {
	if base == nil || base.pool == nil || base.coordinator == nil ||
		!base.security.Valid() || base.platformIAM == nil || base.epoch == nil ||
		base.clock == nil {
		return nil, fmt.Errorf("IAM platform access dependencies are required")
	}
	return &PlatformAccessRuntime{base: base}, nil
}
