package documents

import (
	adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/v1adoption"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

func CutoverApplyContract() (platformcutover.ApplyContract, error) {
	descriptor, err := adoptionpostgres.NewCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}
