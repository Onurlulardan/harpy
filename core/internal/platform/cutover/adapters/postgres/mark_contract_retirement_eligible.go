package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

func (database *ContractRetirementDatabase) MarkEligible(
	ctx context.Context, key contractretirement.Key,
	identity contractretirement.Identity, expected contractretirement.Revision,
) (contractretirement.Revision, error) {
	return database.contractRetirementTransition(
		ctx, markContractRetirementEligibleSQL, key, identity, expected,
	)
}
