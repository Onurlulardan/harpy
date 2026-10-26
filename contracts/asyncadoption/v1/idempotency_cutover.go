package asyncadoptionv1

type IdempotencyCutoverDecision struct {
	sourceContract string
	disposition    string
	reuseInV2      bool
}

func ClassifyV1Idempotency(
	document []byte,
) (IdempotencyCutoverDecision, error) {
	_, err := DecodeV1IdempotencyResponse(document)
	if err != nil {
		return IdempotencyCutoverDecision{}, err
	}
	return IdempotencyCutoverDecision{
		sourceContract: V1IdempotencyContract,
		disposition:    V1IdempotencyDisposition,
		reuseInV2:      false,
	}, nil
}

func (value IdempotencyCutoverDecision) SourceContract() string {
	return value.sourceContract
}

func (value IdempotencyCutoverDecision) Disposition() string {
	return value.disposition
}

func (value IdempotencyCutoverDecision) ReuseInV2() bool {
	return value.reuseInV2
}
