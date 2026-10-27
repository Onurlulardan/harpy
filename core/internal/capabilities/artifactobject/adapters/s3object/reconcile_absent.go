package s3object

import (
	"context"
	"errors"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) reconcileAbsentHead(
	ctx context.Context,
	expected artifactobject.ReconcileExpected,
	before reconciliationSnapshot,
	headErr error,
) (artifactobject.ReconcileResult, error) {
	if before.present || !provesAbsence(headErr) {
		return unknownReconciliation(headErr)
	}
	requestID := errorRequestID(headErr)
	if requestID == "" {
		return unknownReconciliation(errors.Join(headErr, ErrProviderEvidence))
	}
	after, err := store.listReconciliationSnapshot(ctx, expected.Key)
	if err != nil || !after.sameObjectState(before) {
		return unknownReconciliation(errors.Join(err, headErr,
			ErrProviderEvidence))
	}
	verifiedAt, err := store.evidenceNow()
	if err != nil {
		return unknownReconciliation(err)
	}
	return artifactobject.ReconcileResult{
		State:             artifactobject.ReconciliationMissing,
		ProviderRequestID: requestID,
		EvidenceSHA256: reconcileEvidenceSHA(expected,
			before, after, artifactobject.ReconciliationMissing, requestID),
		VerifiedAt: verifiedAt,
	}, nil
}

func unknownReconciliation(err error) (artifactobject.ReconcileResult, error) {
	return artifactobject.ReconcileResult{State: artifactobject.ReconciliationUnknown},
		errors.Join(artifactobject.ErrOutcomeUnknown, err)
}
