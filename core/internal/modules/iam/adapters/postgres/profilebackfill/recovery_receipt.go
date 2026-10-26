package profilebackfill

import (
	"context"
	"crypto/sha256"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func (reader *RecoveryReader) LookupReceipt(
	ctx context.Context,
	scope tenantScope,
	evidence profilebackfillapp.Evidence,
) (profilebackfillapp.RecoveryReceipt, bool, error) {
	if !reader.usable() || evidence.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != evidence.ActorID {
		return profilebackfillapp.RecoveryReceipt{}, false,
			profilebackfillapp.ErrInvalid
	}
	bound, err := reader.value.bind(scope)
	if err != nil {
		return profilebackfillapp.RecoveryReceipt{}, false, err
	}
	return readRecoveryReceipt(ctx, bound, evidence)
}

func readRecoveryReceipt(
	ctx context.Context,
	bound *boundStore,
	evidence profilebackfillapp.Evidence,
) (profilebackfillapp.RecoveryReceipt, bool, error) {
	var receipt profilebackfillapp.RecoveryReceipt
	var request []byte
	err := bound.database.QueryRow(
		ctx, readRecoveryReceiptSQL, evidence.OperationID, evidence.ActorID,
		evidence.ManagePermissionID, evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	).Scan(
		&request, &receipt.PlanID, &receipt.RecoveryOfPlanID,
		&receipt.RecoveryRootPlanID, &receipt.RecoverySequence,
		&receipt.Revision, &receipt.State, &receipt.AuthorizationClock,
	)
	if receiptMissing(err) {
		return profilebackfillapp.RecoveryReceipt{}, false, nil
	}
	if err != nil || len(request) != sha256.Size {
		return profilebackfillapp.RecoveryReceipt{}, false,
			profilebackfillapp.ErrUnavailable
	}
	copy(receipt.RequestSHA256[:], request)
	clear(request)
	if receipt.Validate() != nil {
		return profilebackfillapp.RecoveryReceipt{}, false,
			profilebackfillapp.ErrUnavailable
	}
	return receipt, true, nil
}
