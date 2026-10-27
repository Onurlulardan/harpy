package iam

import "fmt"

type AuditAccessRuntime struct{ base *BillingAccessRuntime }

func NewAuditAccessRuntime(base *BillingAccessRuntime) (*AuditAccessRuntime, error) {
	if base == nil || base.pool == nil || base.coordinator == nil ||
		!base.security.Valid() || base.epoch == nil ||
		base.clock == nil {
		return nil, fmt.Errorf("IAM Audit access dependencies are required")
	}
	return &AuditAccessRuntime{base: base}, nil
}
