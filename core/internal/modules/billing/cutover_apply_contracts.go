package billing

import (
	v1adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/v1adoption"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

func GlobalCutoverApplyContract() (platformcutover.ApplyContract, error) {
	descriptor, err := v1adoptionpostgres.NewGlobalCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}

func TenantCutoverApplyContract() (platformcutover.ApplyContract, error) {
	descriptor, err := v1adoptionpostgres.NewTenantCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}

func V1BaseAdoptionContract() (platformcutover.ApplyContract, error) {
	descriptor, err := v1adoptionpostgres.NewBaseCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}
