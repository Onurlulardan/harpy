package iam

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func normalizeHTTPLogin(
	input AuthHTTPLoginInput,
) (identifier.Value, platformtx.CredentialIdentifierKind, authentication.ClientID, error) {
	switch input.Mode {
	case AuthHTTPLoginOperator:
		value, err := identifier.NormalizeEmailV1(input.Identifier)
		return value, platformtx.CredentialEmailIdentifier,
			authentication.ClientOperatorConsole, err
	case AuthHTTPLoginPortal:
		value, err := identifier.NormalizePortalV1(input.Identifier)
		return value, platformtx.CredentialPortalIdentifier,
			authentication.ClientPortal, err
	default:
		return identifier.Value{}, "", "", fmt.Errorf("login mode is invalid")
	}
}
