package recoveryevidence

import (
	"context"
	"fmt"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports/recoveryevidence"
)

func (store *Store) ReadRecoveryHead(
	ctx context.Context,
	environmentID string,
) (evidenceports.Head, error) {
	var result evidenceports.Head
	err := store.within(ctx, func(database database) error {
		return database.QueryRow(ctx, readHeadSQL, environmentID).Scan(
			&result.BindingID, &result.EnvironmentID,
			&result.DeploymentMode, &result.RegistryKind,
			&result.LicenseBindingID, &result.FinalizationReceiptSHA256,
			&result.RecoveryTrustBundleSHA256, &result.SignerGeneration,
			&result.SignerKID, &result.SignerPublicKeySHA256,
			&result.BindingProofSHA256, &result.BindingProofBytes,
			&result.OperationID, &result.Purpose, &result.ScopeKind,
			&result.ScopeID, &result.Generation, &result.PreviousHeadSHA256,
			&result.StateSHA256, &result.StateBytes,
			&result.ProposalSHA256, &result.ProposalBytes,
			&result.AckSHA256, &result.AckBytes,
			&result.RegistrySequence, &result.CompletedAt,
		)
	})
	if err != nil {
		return evidenceports.Head{},
			fmt.Errorf("read protected authority recovery evidence: %w", err)
	}
	return result, nil
}

var _ evidenceports.Reader = (*Store)(nil)
