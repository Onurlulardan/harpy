package resourceentitlement

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"

var (
	ErrConfiguration     = billingapi.ErrMeteredResourceConfiguration
	ErrRequest           = billingapi.ErrMeteredResourceRequest
	ErrUnavailable       = billingapi.ErrMeteredResourceUnavailable
	ErrIdempotencyReused = billingapi.ErrMeteredResourceIdempotencyReused
	ErrResourceState     = billingapi.ErrMeteredResourceStateConflict
	ErrEvidence          = billingapi.ErrMeteredResourceEvidence
)
