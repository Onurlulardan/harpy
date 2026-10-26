package manual

import (
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

var receiptNamespace = uuid.MustParse("add006c1-fb1c-50ce-9c51-392d296b9dcb")

func (adapter *Adapter) receipt(
	request effectsv1.Request,
	operation string,
	code string,
) (providerapp.EffectResult, error) {
	if adapter == nil || adapter.clock == nil || request.Validate() != nil ||
		request.Provider != string(adapter.Descriptor().Key()) ||
		request.Operation != operation {
		return providerapp.EffectResult{}, providerapp.ErrEffectResult
	}
	completedAt := adapter.clock.Now()
	if completedAt.IsZero() || completedAt.Nanosecond()%int(time.Microsecond) != 0 {
		return providerapp.EffectResult{}, providerapp.ErrProviderDependency
	}
	completedAt = time.UnixMicro(completedAt.UnixMicro()).UTC()
	receiptID := uuid.NewSHA1(
		receiptNamespace, []byte(request.EffectID+"\x00"+operation),
	)
	providerReference := "manual/" + request.EffectID
	receipt, err := effectsv1.NewSucceededReceipt(
		request, receiptID.String(), code, &providerReference,
		request.PayloadReference, request.PayloadSHA256,
		completedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return providerapp.EffectResult{}, providerapp.ErrEffectResult
	}
	return providerapp.DefinitiveEffect(request, receipt)
}
