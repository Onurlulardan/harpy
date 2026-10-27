package adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (worker *AdoptionCopyWorker) Copy(ctx context.Context, runner JobRunner,
	authority Authority, sourceVersionID domain.ID,
) (Checkpoint, error) {
	if worker == nil || foundation.NilPort(runner) || ctx == nil || ctx.Err() != nil ||
		!platformtx.NetworkAllowed(ctx) || !validAuthority(authority) ||
		!sourceVersionID.UUIDv7() {
		return Checkpoint{}, ErrInvalid
	}
	plan, err := worker.prepare(ctx, runner, authority, sourceVersionID)
	if err != nil {
		return Checkpoint{}, err
	}
	receipt, err := worker.copier.Copy(ctx, authority, plan)
	if err != nil {
		return Checkpoint{}, mapError(err)
	}
	receipt.VerifiedAt = worker.clock.Now().UTC()
	if !validCopyReceipt(authority, plan, receipt) {
		return Checkpoint{}, ErrUnavailable
	}
	return worker.finalize(ctx, runner, authority, plan, receipt)
}

func validCopyReceipt(authority Authority, plan Checkpoint, receipt CopyReceipt) bool {
	expected, err := domain.ContentObjectKey(authority.TenantID,
		plan.TargetItemID, plan.TargetVersionID)
	return err == nil && receipt.TargetObject.Key == expected &&
		receipt.TargetObject.Version != "" && receipt.TargetObject.Generation > 0 &&
		receipt.TargetObject.PlaintextBytes == plan.SourceBytes &&
		receipt.TargetObject.PlaintextSHA256 == plan.SourceSHA256 &&
		receipt.TargetObject.CiphertextBytes > plan.SourceBytes &&
		receipt.TargetObject.CiphertextSHA256 != [32]byte{} &&
		receipt.TargetObject.KeyID != "" && len(receipt.TargetObject.WrappedDEK) >= 115 &&
		receipt.MetadataSHA256 == plan.SourceMetadataSHA256 &&
		validLSN(receipt.AppliedLSN) && !receipt.VerifiedAt.IsZero()
}
