package profilepostgres

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"

type profileNormalizer struct{}

func (profileNormalizer) NormalizeEmailV1(value string) (string, error) {
	normalized, err := identifier.NormalizeEmailV1(value)
	if err != nil {
		return "", err
	}
	return normalized.Normalized(), nil
}

func (profileNormalizer) NormalizePortalV1(value string) (string, error) {
	normalized, err := identifier.NormalizePortalV1(value)
	if err != nil {
		return "", err
	}
	return normalized.Normalized(), nil
}
