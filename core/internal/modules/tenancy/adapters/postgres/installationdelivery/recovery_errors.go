package installationdelivery

import tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"

var (
	errRecoveryUnavailable = tenantapi.ErrOnboardingUnavailable
	errRecoveryConflict    = tenantapi.ErrOnboardingConflict
)
