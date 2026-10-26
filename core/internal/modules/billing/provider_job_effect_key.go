package billing

import (
	"regexp"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

var providerReconcileEffectKeyPattern = regexp.MustCompile(
	`^billing:provider:reconcile:[0-9a-f]{64}$`,
)

func validProviderEffectKey(
	kind string,
	provider string,
	referenceID uuid.UUID,
	effectKey string,
) bool {
	switch kind {
	case providerapp.ProviderCommandJobKind:
		return effectKey == "billing:provider:command:"+
			provider+":"+referenceID.String()
	case providerapp.ProviderReconciliationJobKind:
		return providerReconcileEffectKeyPattern.MatchString(effectKey)
	default:
		return false
	}
}
