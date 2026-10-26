package iam

import (
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformactorbridge "github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
)

type ActorBridgeIssuerIAM struct {
	authorizer authorization.Authorizer
	valid      bool
}

func NewActorBridgeIssuerIAM() (*ActorBridgeIssuerIAM, error) {
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM actor bridge evidence adapter: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM actor bridge authorizer: %w", err)
	}
	return &ActorBridgeIssuerIAM{authorizer: authorizer, valid: true}, nil
}

func (runtime *ActorBridgeIssuerIAM) Valid() bool {
	return runtime != nil && runtime.valid
}

var _ platformactorbridge.AuthorizationPort = (*ActorBridgeIssuerIAM)(nil)
