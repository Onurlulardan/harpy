package recoveryorchestration

import (
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	iamapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api"
	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type RecoveryFactories struct {
	Documents func(
		any, documentsapi.RecoveryIdentity,
	) (documentsapi.Recovery, error)
	Targets func(
		any, string,
	) (tenantapi.RecoveryTargets, error)
	IAMEvidence func(
		any,
	) (iamapi.RecoveryEvidence, error)
	AuthorityEvidence func(
		any,
	) (authorityapi.RecoveryEvidence, error)
}

func (factories RecoveryFactories) Valid() bool {
	return factories.Documents != nil && factories.Targets != nil &&
		factories.IAMEvidence != nil && factories.AuthorityEvidence != nil
}
