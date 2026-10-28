package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
)

func (executor *TenantExecutor) Migrate(
	ctx context.Context, claim tenantexec.Claim,
) (tenantexec.MigrationReceipt, error) {
	if err := executor.validateClaim(claim); err != nil {
		return tenantexec.MigrationReceipt{}, err
	}
	if err := executor.startRun(ctx, claim); err != nil {
		return tenantexec.MigrationReceipt{}, err
	}
	if err := executor.ensureSchema(ctx, claim); err != nil {
		executor.failRun(context.WithoutCancel(ctx), claim, "schema_provision_failed")
		return tenantexec.MigrationReceipt{}, err
	}
	versions, err := executor.target.Execute(
		ctx, claim.TenantID.String(), claim.SchemaName,
		claim.MigrationRunID.String(),
	)
	if err != nil && !errors.Is(err, ErrCommitted) {
		// #region agent log
		fmt.Fprintf(os.Stderr, "debug J %s\n", strings.ReplaceAll(err.Error(), "\n", " "))
		// #endregion
		if errors.Is(err, ErrOutcomeAmbiguous) {
			return tenantexec.MigrationReceipt{}, errors.Join(
				tenantexec.ErrAmbiguousOutcome, err,
			)
		}
		executor.failRun(context.WithoutCancel(ctx), claim, "migration_execution_failed")
		return tenantexec.MigrationReceipt{}, err
	}
	if err := executor.completeRun(context.WithoutCancel(ctx), claim); err != nil {
		return tenantexec.MigrationReceipt{}, err
	}
	evidence := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%s:%s:%v", claim.TenantID, claim.MigrationRunID,
		claim.MigrationManifestSHA256, versions,
	)))
	evidenceDigest := hex.EncodeToString(evidence[:])
	receipt := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%d:%s", claim.JobID, claim.Attempt, evidenceDigest,
	)))
	return tenantexec.MigrationReceipt{
		ReceiptSHA256:  hex.EncodeToString(receipt[:]),
		EvidenceSHA256: evidenceDigest,
	}, nil
}

func (executor *TenantExecutor) validateClaim(claim tenantexec.Claim) error {
	release := executor.release
	if claim.Origin != "new" || claim.ReleaseID.String() != release.ReleaseID ||
		claim.DeploymentMode == "" ||
		claim.RuntimeGeneration != release.RuntimeGeneration ||
		claim.MigrationManifestSHA256 != release.MigrationManifestSHA256 ||
		claim.JobID.String() == "" || claim.SchemaName == "" {
		return fmt.Errorf("tenant migration claim differs from verified release")
	}
	return nil
}

var _ tenantexec.Migrator = (*TenantExecutor)(nil)
