package operator

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func reconcileRestoredIdentity(ctx context.Context) (
	identityReconcileOutcome,
	string,
	error,
) {
	settings, err := config.LoadOperatorIdentityReconcile()
	if err != nil {
		return identityReconcileOutcome{}, "configuration", err
	}
	if _, err = buildinfo.ReadOperator(settings.Runtime().Environment()); err != nil {
		return identityReconcileOutcome{}, "build_contract", err
	}
	artifacts, err := loadRestoredIdentityArtifacts(settings, time.Now().UTC())
	if err != nil {
		return identityReconcileOutcome{}, "evidence", err
	}
	rotation, err := artifacts.custody.RotateForRestore(artifacts.restoreOperation)
	if err != nil {
		return identityReconcileOutcome{}, "custody", err
	}
	custodyDigest, err := decodeDigest(rotation.ResultSHA256)
	if err != nil {
		return identityReconcileOutcome{}, "custody", err
	}
	epochDigest := sha256.Sum256(rotation.SecurityEpoch[:])
	receipt, err := iammodule.BuildRestoredIdentityReceipt(
		iammodule.RestoredIdentityReceiptInput{
			RestoreOperationID:   artifacts.restoreOperation,
			EnvironmentID:        artifacts.environment,
			RecoverySetID:        artifacts.recoverySet,
			TenantID:             artifacts.tenant,
			OwnerUserID:          artifacts.owner,
			AnchorID:             artifacts.anchor,
			NewDatabaseSystemID:  artifacts.newDatabase,
			SecurityEvidence:     artifacts.security,
			RestoreReceiptSHA256: artifacts.receiptSHA256,
			SecurityEpochSHA256:  epochDigest,
			Custody:              rotation,
			OccurredAt:           artifacts.occurredAt,
		},
	)
	if err != nil {
		return identityReconcileOutcome{}, "evidence", err
	}
	pool, err := openIdentityReconcilePool(ctx, settings.Runtime())
	if err != nil {
		return identityReconcileOutcome{}, "database", err
	}
	defer pool.Close()
	store, err := iammodule.NewRestoredIdentityReconciler(pool)
	if err != nil {
		return identityReconcileOutcome{}, "composition", err
	}
	result, err := store.Reconcile(ctx, iammodule.RestoredIdentityCommand{
		RestoreOperationID: artifacts.restoreOperation, EnvironmentID: artifacts.environment,
		RecoverySetID: artifacts.recoverySet, TenantID: artifacts.tenant,
		OwnerUserID: artifacts.owner, AnchorID: artifacts.anchor,
		NewDatabaseSystemID: artifacts.newDatabase, SecurityEpoch: rotation.SecurityEpoch,
		RestoreReceiptSHA256: artifacts.receiptSHA256, ProtectedHeadSHA256: artifacts.protectedHead,
		CustodyGeneration: rotation.Generation, CustodyResultSHA256: custodyDigest,
		Receipt: receipt, OccurredAt: artifacts.occurredAt,
	})
	if err != nil {
		return identityReconcileOutcome{}, "reconcile", err
	}
	if result.State != iammodule.RestoredIdentityStateReactivationRequired {
		return identityReconcileOutcome{}, "reconcile", fmt.Errorf("invalid admission state")
	}
	return identityReconcileOutcome{OperationID: artifacts.restoreOperation,
		EnvironmentID: artifacts.environment, TenantID: artifacts.tenant,
		State: result.State, Replayed: result.Replayed,
		InvalidationSHA256: result.InvalidationSHA256,
		CustodyGeneration:  rotation.Generation, CustodySHA256: custodyDigest}, "reconcile", nil
}
