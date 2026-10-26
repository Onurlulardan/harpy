package maintenance

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"

type maintenanceProfileNormalizer struct{}

func (maintenanceProfileNormalizer) NormalizeEmailV1(value string) (string, error) {
	normalized, err := identifier.NormalizeEmailV1(value)
	if err != nil {
		return "", err
	}
	return normalized.Normalized(), nil
}

func (maintenanceProfileNormalizer) NormalizePortalV1(value string) (string, error) {
	normalized, err := identifier.NormalizePortalV1(value)
	if err != nil {
		return "", err
	}
	return normalized.Normalized(), nil
}
