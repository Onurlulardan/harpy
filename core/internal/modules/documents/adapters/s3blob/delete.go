package s3blob

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func (store *Store) DeleteExact(ctx context.Context,
	task deletion.Task,
) (deletion.DeleteReceipt, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !task.ID.UUIDv7() ||
		!validKey(task.ObjectKey) || task.ObjectVersion == "" ||
		task.ObjectGeneration == 0 || task.CiphertextBytes < 1 {
		return deletion.DeleteReceipt{}, deletion.ErrInvalid
	}
	output, err := store.api.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &store.bucket,
		Key: &task.ObjectKey, VersionId: &task.ObjectVersion,
		ExpectedBucketOwner: store.owner})
	providerID := ""
	status := "deleted"
	if err != nil {
		if !provesAbsence(err) {
			return deletion.DeleteReceipt{}, errors.Join(deletion.ErrUnavailable, err)
		}
		providerID, status = errorRequestID(err), "already_absent"
	} else if output != nil {
		providerID = requestID(output.ResultMetadata)
	}
	if providerID == "" {
		return deletion.DeleteReceipt{}, errors.Join(deletion.ErrUnavailable, ErrIntegrity)
	}
	deleteDigest := foundation.Digest([]byte("documents-object-delete.v1"),
		[]byte(task.ObjectKey), []byte(task.ObjectVersion), []byte(providerID), []byte(status))
	absenceDigest, verifiedAt, err := store.probeAbsent(ctx, task)
	if err != nil {
		return deletion.DeleteReceipt{}, err
	}
	return deletion.DeleteReceipt{ProviderRequestID: providerID,
		DeleteSHA256: deleteDigest, AbsenceSHA256: absenceDigest,
		VerifiedAt: verifiedAt}, nil
}

func (store *Store) probeAbsent(ctx context.Context,
	task deletion.Task,
) ([32]byte, time.Time, error) {
	_, err := store.api.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &store.bucket,
		Key: &task.ObjectKey, VersionId: &task.ObjectVersion,
		ExpectedBucketOwner: store.owner})
	if err == nil || !provesAbsence(err) {
		return [32]byte{}, time.Time{}, errors.Join(deletion.ErrUnavailable, ErrAmbiguous, err)
	}
	providerID := errorRequestID(err)
	if providerID == "" {
		return [32]byte{}, time.Time{}, errors.Join(deletion.ErrUnavailable, ErrIntegrity)
	}
	verifiedAt := store.clock.Now().UTC()
	digest := foundation.Digest([]byte("documents-object-absence.v1"),
		[]byte(task.ObjectKey), []byte(task.ObjectVersion), []byte(providerID),
		[]byte(strconv.FormatUint(task.ObjectGeneration, 10)))
	return digest, verifiedAt, nil
}

var _ deletion.ObjectStore = (*Store)(nil)
