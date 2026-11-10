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
	if input.Mode != AuthHTTPLoginOperator {
		return identifier.Value{}, "", "", fmt.Errorf("login mode is invalid")
	}
	value, err := identifier.NormalizeEmailV1(input.Identifier)
	return value, platformtx.CredentialEmailIdentifier,
		authentication.ClientOperatorConsole, err
}
