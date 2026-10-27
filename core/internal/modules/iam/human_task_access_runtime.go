package iam

import (
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	humantaskpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/humantask"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	humantaskapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/humantask"
)

type HumanTaskAccessRuntime struct{ verifier humantaskapp.Verifier }

func NewHumanTaskAccessRuntime() (*HumanTaskAccessRuntime, error) {
	authorizationPort, err := authorizationpostgres.NewWorkflowHumanTaskAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build Workflow human task authorization port: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(authorizationPort)
	if err != nil {
		return nil, fmt.Errorf("build Workflow human task authorizer: %w", err)
	}
	candidates, err := humantaskpostgres.NewCandidateReader()
	if err != nil {
		return nil, fmt.Errorf("build Workflow human task candidate port: %w", err)
	}
	verifier, err := humantaskapp.NewVerifier(
		humanTaskCurrentPermission{authorizer: authorizer}, candidates, candidates,
	)
	if err != nil {
		return nil, fmt.Errorf("build Workflow human task verifier: %w", err)
	}
	return &HumanTaskAccessRuntime{verifier: verifier}, nil
}

func (runtime *HumanTaskAccessRuntime) Valid() bool { return runtime != nil }
