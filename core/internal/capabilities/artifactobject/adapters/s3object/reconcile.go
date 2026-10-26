package s3object

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) ReconcileExact(
	ctx context.Context,
	expected artifactobject.ReconcileExpected,
) (artifactobject.ReconcileResult, error) {
	if err := store.reconciliationInputError(ctx, expected); err != nil {
		return artifactobject.ReconcileResult{}, err
	}
	before, err := store.listReconciliationSnapshot(ctx, expected.Key)
	if err != nil {
		return unknownReconciliation(err)
	}
	head, headErr := store.api.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &store.bucket, Key: &expected.Key,
		ExpectedBucketOwner: store.owner, ChecksumMode: types.ChecksumModeEnabled,
		VersionId: optionalVersion(before),
	})
	if headErr != nil {
		return store.reconcileAbsentHead(ctx, expected, before, headErr)
	}
	if !before.present {
		return unknownReconciliation(ErrProviderEvidence)
	}
	evidence, err := store.verifyReconciliationHead(head, expected, before)
	if err != nil {
		return unknownReconciliation(err)
	}
	after, err := store.listReconciliationSnapshot(ctx, expected.Key)
	if err != nil || !after.sameObjectState(before) {
		return unknownReconciliation(errors.Join(err, ErrProviderEvidence))
	}
	verifiedAt, err := store.evidenceNow()
	if err != nil {
		return unknownReconciliation(err)
	}
	return artifactobject.ReconcileResult{State: artifactobject.ReconciliationSealed,
		Identity:          artifactobject.Identity{Key: expected.Key, Version: before.version},
		ProviderRequestID: evidence.requestID,
		EvidenceSHA256: reconcileEvidenceSHA(expected, before, after,
			artifactobject.ReconciliationSealed, evidence.requestID),
		ObjectModifiedAt: evidence.modifiedAt,
		VerifiedAt:       verifiedAt}, nil
}

func (store *Store) reconciliationInputError(
	ctx context.Context,
	expected artifactobject.ReconcileExpected,
) error {
	if err := store.operationError(ctx); err != nil {
		return err
	}
	if !store.validObjectKey(expected.Key) || expected.CiphertextBytes < 1 ||
		expected.CiphertextBytes > store.maximumBytes ||
		expected.CiphertextSHA256 == [32]byte{} || expected.HeaderSHA256 == [32]byte{} {
		return ErrObjectInput
	}
	return nil
}

func optionalVersion(snapshot reconciliationSnapshot) *string {
	if !snapshot.present {
		return nil
	}
	return &snapshot.version
}
