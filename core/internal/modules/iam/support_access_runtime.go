package iam

import "fmt"

type SupportAccessRuntime struct{ base *BillingAccessRuntime }

func NewSupportAccessRuntime(base *BillingAccessRuntime) (*SupportAccessRuntime, error) {
	if base == nil || base.pool == nil || base.coordinator == nil ||
		!base.security.Valid() || base.platformIAM == nil || base.epoch == nil ||
		base.sessions == nil || base.clock == nil {
		return nil, fmt.Errorf("IAM Support access dependencies are required")
	}
	return &SupportAccessRuntime{base: base}, nil
}
