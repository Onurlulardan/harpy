package usage

type EffectEvidence string

const (
	EffectUnknown     EffectEvidence = "unknown"
	EffectConfirmed   EffectEvidence = "effect_confirmed"
	NoEffectConfirmed EffectEvidence = "no_effect_confirmed"
)

func (evidence EffectEvidence) Valid() bool {
	return evidence == EffectUnknown || evidence == EffectConfirmed ||
		evidence == NoEffectConfirmed
}

func ResolveExpiry(effectStarted bool, evidence EffectEvidence) (ReservationState, error) {
	if !evidence.Valid() {
		return "", ErrExpiry
	}
	if !effectStarted {
		if evidence == EffectConfirmed {
			return "", ErrExpiry
		}
		return Released, nil
	}
	switch evidence {
	case EffectUnknown:
		return ReconcileRequired, nil
	case EffectConfirmed:
		return Committed, nil
	case NoEffectConfirmed:
		return Released, nil
	default:
		return "", ErrExpiry
	}
}
