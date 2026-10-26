package catalog

import (
	adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/adoption"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

// CutoverApplyContract exposes Catalog's only production V1 adoption write path.
func CutoverApplyContract() (platformcutover.ApplyContract, error) {
	descriptor, err := adoptionpostgres.NewCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}
