package profile

type TextNormalizer interface {
	NormalizeEmailV1(string) (string, error)
	NormalizePortalV1(string) (string, error)
}

func normalizeProfileText(
	field Field, value string, normalizer TextNormalizer,
) (string, error) {
	switch field.normalization {
	case NormalizationNoneV1:
		return value, nil
	case NormalizationEmailV1:
		if normalizer == nil {
			return "", ErrValidationInvalid
		}
		return normalizer.NormalizeEmailV1(value)
	case NormalizationPortalV1:
		if normalizer == nil {
			return "", ErrValidationInvalid
		}
		return normalizer.NormalizePortalV1(value)
	default:
		return "", ErrValidationInvalid
	}
}
