package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *WorkerStore) CompleteDeleted(
	ctx context.Context, claim exportapp.DeleteClaim,
	receipt exportapp.DeleteReceipt, completedAt time.Time,
) error {
	if claim.TenantID == uuid.Nil || claim.ExportID.Version() != 7 ||
		claim.DeleteJobID.Version() != 7 || claim.LeaseOwner == uuid.Nil ||
		claim.DeleteGeneration < 1 || claim.DeleteGeneration > exportapp.MaximumAttempts ||
		receipt.VerifiedAbsentAt.IsZero() || completedAt.IsZero() ||
		completedAt.Before(receipt.VerifiedAbsentAt) {
		return exportapp.ErrInvalid
	}
	if claim.Draft != nil && (receipt.ProviderRequestID == "" ||
		receipt.AbsenceSHA256 == [32]byte{}) {
		return exportapp.ErrInvalid
	}
	if claim.ObjectVersion != "" && receipt.DeleteResponseSHA256 == [32]byte{} {
		return exportapp.ErrInvalid
	}
	var persisted bool
	err := storage.inTenant(ctx, claim.TenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		return bound.queryRow(boundContext, workerDeleteCompleteSQL,
			claim.ExportID, claim.DeleteJobID, claim.LeaseOwner, claim.DeleteGeneration,
			optionalText(claim.ObjectVersion), optionalText(receipt.ProviderRequestID),
			optionalDigest(receipt.DeleteResponseSHA256), optionalDigest(receipt.AbsenceSHA256),
			receipt.VerifiedAbsentAt, completedAt).Scan(&persisted)
	})
	if err != nil {
		return mapError(err, nil)
	}
	if !persisted {
		return exportapp.ErrUnavailable
	}
	return nil
}

func optionalText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func optionalDigest(value [32]byte) any {
	if value == [32]byte{} {
		return nil
	}
	return value[:]
}
