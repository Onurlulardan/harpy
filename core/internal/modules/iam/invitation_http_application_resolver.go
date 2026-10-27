package iam

import (
	"fmt"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func newInvitationResolver() (platformtx.CredentialResolverBinder, error) {
	resolver, err := iampostgres.NewCredentialTargetResolverBinder()
	if err != nil {
		return platformtx.CredentialResolverBinder{}, fmt.Errorf(
			"build IAM invitation target resolver: %w", err,
		)
	}
	return resolver, nil
}
