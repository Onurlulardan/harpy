package iam

import (
	identifierapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	"github.com/google/uuid"
)

type UUID = uuid.UUID

func NormalizeEmailV1(value string) (string, error) {
	normalized, err := identifierapp.NormalizeEmailV1(value)
	if err != nil {
		return "", err
	}
	return normalized.Normalized(), nil
}
