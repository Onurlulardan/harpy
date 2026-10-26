package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

func (database *ContractRetirementDatabase) contractRetirementTransition(
	ctx context.Context, query string, key contractretirement.Key,
	identity contractretirement.Identity, expected contractretirement.Revision,
) (contractretirement.Revision, error) {
	keyValue, owner, domain, err := contractRetirementIdentityArguments(key, identity)
	if err != nil {
		return contractretirement.Revision{}, err
	}
	expectedValue, err := expected.DatabaseValue()
	if err != nil {
		return contractretirement.Revision{}, err
	}
	var revisionValue int64
	err = database.queryer.QueryRow(ctx, query,
		keyValue, owner, domain, expectedValue).Scan(&revisionValue)
	return contractRetirementRevision(revisionValue, err)
}

func contractRetirementRevision(
	value int64, err error,
) (contractretirement.Revision, error) {
	if err != nil {
		return contractretirement.Revision{}, classifyContractRetirementError(err)
	}
	revision, err := contractretirement.NewRevision(value)
	if err != nil {
		return contractretirement.Revision{}, contractretirement.NewFault(
			contractretirement.FaultDependency,
		)
	}
	return revision, nil
}
