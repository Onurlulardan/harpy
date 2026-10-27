package entitlement

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"

// This root-composition alias exposes Billing's public owner contract without
// defining a second Analytics entitlement model or validation policy.
type MutationDecider = billingapi.MeteredResourceMutationDecider
