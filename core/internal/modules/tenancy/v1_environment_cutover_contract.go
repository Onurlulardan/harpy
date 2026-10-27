package tenancy

import (
	v1environmentpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/v1environment"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

func V1EnvironmentAdoptionContract() (platformcutover.ApplyContract, error) {
	descriptor, err := v1environmentpostgres.NewCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}
