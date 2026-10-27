package iam

import (
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
)

func newSystemLayoutAuthorizer() (authorization.Authorizer, error) {
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return authorization.Authorizer{}, fmt.Errorf("build IAM system layout authorization evidence: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return authorization.Authorizer{}, fmt.Errorf("build IAM system layout authorizer: %w", err)
	}
	return authorizer, nil
}
