package providerapp

import effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"

type EffectResult struct {
	receipt effectsv1.Receipt
	unknown bool
}

func DefinitiveEffect(
	request effectsv1.Request,
	receipt effectsv1.Receipt,
) (EffectResult, error) {
	if request.Validate() != nil || receipt.ValidateForRequest(request) != nil {
		return EffectResult{}, ErrEffectResult
	}
	return EffectResult{receipt: cloneReceipt(receipt)}, nil
}

func UnknownEffect(request effectsv1.Request) (EffectResult, error) {
	if request.Validate() != nil {
		return EffectResult{}, ErrEffectResult
	}
	return EffectResult{unknown: true}, nil
}

func (result EffectResult) IsUnknown() bool { return result.unknown }

func (result EffectResult) Receipt() (effectsv1.Receipt, bool) {
	if result.unknown || result.receipt.Validate() != nil {
		return effectsv1.Receipt{}, false
	}
	return cloneReceipt(result.receipt), true
}

func (result EffectResult) ValidFor(request effectsv1.Request) bool {
	if request.Validate() != nil {
		return false
	}
	if result.unknown {
		return result.receipt == (effectsv1.Receipt{})
	}
	return result.receipt.ValidateForRequest(request) == nil
}

func cloneReceipt(receipt effectsv1.Receipt) effectsv1.Receipt {
	document, err := receipt.MarshalJSON()
	if err != nil {
		return effectsv1.Receipt{}
	}
	copy, err := effectsv1.DecodeReceipt(document)
	if err != nil {
		return effectsv1.Receipt{}
	}
	return copy
}
