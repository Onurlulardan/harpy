package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

func (database *ContractRetirementDatabase) AdmitManifest(
	ctx context.Context, batch contractretirement.Batch,
) (contractretirement.Decision, error) {
	return database.retirementDecision(ctx, admitContractRetirementManifestSQL, batch)
}
