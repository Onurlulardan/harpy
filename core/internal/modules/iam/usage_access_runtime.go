package iam

import (
	"fmt"

	"github.com/google/uuid"
)

type UsageAccessRuntime struct {
	base          *BillingAccessRuntime
	environmentID uuid.UUID
}

func NewUsageAccessRuntime(
	base *BillingAccessRuntime, environmentID uuid.UUID,
) (*UsageAccessRuntime, error) {
	if base == nil || base.pool == nil || base.coordinator == nil ||
		!base.security.Valid() || base.platformIAM == nil || base.epoch == nil ||
		base.sessions == nil || base.clock == nil || environmentID == uuid.Nil {
		return nil, fmt.Errorf("IAM Usage access dependencies are required")
	}
	return &UsageAccessRuntime{base: base, environmentID: environmentID}, nil
}
