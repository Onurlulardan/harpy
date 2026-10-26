package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

func (database *ContractRetirementDatabase) Reconcile(
	ctx context.Context, key contractretirement.Key,
	identity contractretirement.Identity, count contractretirement.RemainingScopeCount,
	expected contractretirement.Revision,
) (contractretirement.Revision, error) {
	keyValue, owner, domain, err := contractRetirementIdentityArguments(key, identity)
	if err != nil {
		return contractretirement.Revision{}, err
	}
	countValue, err := count.DatabaseValue()
	if err != nil {
		return contractretirement.Revision{}, err
	}
	expectedValue, err := expected.DatabaseValue()
	if err != nil {
		return contractretirement.Revision{}, err
	}
	var revisionValue int64
	err = database.queryer.QueryRow(ctx, reconcileContractRetirementSQL,
		keyValue, owner, domain, countValue, expectedValue).Scan(&revisionValue)
	return contractRetirementRevision(revisionValue, err)
}
