package billing

import (
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func providerPlatformResult(result providerapp.EffectResult) effectexecution.Result {
	if result.IsUnknown() {
		return effectexecution.Unknown()
	}
	receipt, ok := result.Receipt()
	if !ok {
		return effectexecution.Result{}
	}
	converted, err := effectexecution.Definitive(receipt)
	if err != nil {
		return effectexecution.Result{}
	}
	return converted
}
