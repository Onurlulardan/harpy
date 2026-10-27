package iam

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"

type profileHTTPNormalizer struct{}

func (profileHTTPNormalizer) NormalizeEmailV1(value string) (string, error) {
	normalized, err := identifier.NormalizeEmailV1(value)
	if err != nil {
		return "", err
	}
	return normalized.Normalized(), nil
}

func (profileHTTPNormalizer) NormalizePortalV1(value string) (string, error) {
	normalized, err := identifier.NormalizePortalV1(value)
	if err != nil {
		return "", err
	}
	return normalized.Normalized(), nil
}
