package postrestore

import (
	"context"
	"encoding/hex"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	application "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

func (store *Store) Reconcile(
	ctx context.Context, command application.ReconcileCommand,
) (application.ReconcileResult, error) {
	if store == nil || store.withinTenant == nil || ctx == nil || ctx.Err() != nil {
		return application.ReconcileResult{}, application.ErrUnavailable
	}
	var result application.ReconcileResult
	err := store.tenant(ctx, transactionCommand, command.TenantID,
		func(database database) error {
			var counts json.RawMessage
			var digest []byte
			var tenantReplay bool
			if err := database.QueryRow(ctx, reconcileTenantSQL,
				command.RestoreOperationID, command.EnvironmentID, command.RecoverySetID,
				command.OwnerUserID, command.AnchorID, command.SecurityEpoch[:],
				command.RestoreReceiptSHA256[:], command.ProtectedHeadSHA256[:],
				command.CustodyGeneration, command.CustodyResultSHA256[:], command.OccurredAt,
			).Scan(&counts, &digest, &tenantReplay); err != nil {
				return err
			}
			receipt, err := mergeReceipt(command.Receipt, counts, digest)
			if err != nil {
				return application.ErrUnavailable
			}
			var platformReplay bool
			if err := database.QueryRow(ctx, reconcilePlatformSQL,
				command.RestoreOperationID, command.EnvironmentID, command.RecoverySetID,
				command.TenantID, command.OwnerUserID, command.AnchorID,
				command.NewDatabaseSystemID, command.SecurityEpoch[:],
				command.RestoreReceiptSHA256[:], command.ProtectedHeadSHA256[:],
				command.CustodyGeneration, command.CustodyResultSHA256[:], receipt,
				command.OccurredAt,
			).Scan(&result.State, &platformReplay); err != nil {
				return err
			}
			if len(digest) != 32 {
				return application.ErrUnavailable
			}
			result.InvalidationCounts = append([]byte(nil), counts...)
			copy(result.InvalidationSHA256[:], digest)
			result.Replayed = tenantReplay && platformReplay
			return nil
		})
	return result, err
}

func mergeReceipt(base []byte, counts json.RawMessage, digest []byte) ([]byte, error) {
	if len(digest) != 32 {
		return nil, application.ErrUnavailable
	}
	var receipt map[string]json.RawMessage
	if err := json.Unmarshal(base, &receipt); err != nil || receipt == nil ||
		receipt["invalidation_counts"] != nil || receipt["invalidation_sha256"] != nil {
		return nil, application.ErrUnavailable
	}
	countsCopy, err := json.Marshal(json.RawMessage(counts))
	if err != nil {
		return nil, err
	}
	digestValue, _ := json.Marshal(hex.EncodeToString(digest))
	receipt["invalidation_counts"] = countsCopy
	receipt["invalidation_sha256"] = digestValue
	document, err := json.Marshal(receipt)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(document, 64*1024)
}
