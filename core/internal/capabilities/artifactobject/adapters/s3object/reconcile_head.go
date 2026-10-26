package s3object

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

type reconciliationHead struct {
	requestID  string
	modifiedAt time.Time
}

func (store *Store) verifyReconciliationHead(
	output *s3.HeadObjectOutput,
	expected artifactobject.ReconcileExpected,
	snapshot reconciliationSnapshot,
) (reconciliationHead, error) {
	if output == nil || output.VersionId == nil ||
		*output.VersionId != snapshot.version || output.ContentLength == nil ||
		*output.ContentLength != expected.CiphertextBytes ||
		(output.DeleteMarker != nil && *output.DeleteMarker) ||
		output.LastModified == nil || !validProviderTime(*output.LastModified) ||
		!output.LastModified.Equal(snapshot.lastModified) ||
		snapshot.bytes != expected.CiphertextBytes {
		return reconciliationHead{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, artifactobject.ErrIntegrity,
			ErrProviderEvidence)
	}
	if err := store.verifyMetadata(output.Metadata, output.ChecksumSHA256,
		expected.CiphertextSHA256, expected.HeaderSHA256,
		expected.CiphertextBytes); err != nil {
		return reconciliationHead{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, artifactobject.ErrIntegrity, err)
	}
	requestID, err := metadataRequestID(output.ResultMetadata)
	if err != nil {
		return reconciliationHead{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, err)
	}
	return reconciliationHead{requestID: requestID,
		modifiedAt: output.LastModified.UTC()}, nil
}
