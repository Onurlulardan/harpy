package analytics

import (
	adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/adoption"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

// CutoverApplyContract exposes Analytics' only production V1 adoption write path.
func CutoverApplyContract() (platformcutover.ApplyContract, error) {
	descriptor, err := adoptionpostgres.NewCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}
