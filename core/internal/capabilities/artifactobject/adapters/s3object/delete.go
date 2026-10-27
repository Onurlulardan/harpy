package s3object

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) DeleteExact(ctx context.Context,
	identity artifactobject.Identity) (artifactobject.DeleteResult, error) {
	if err := store.operationError(ctx); err != nil {
		return artifactobject.DeleteResult{}, err
	}
	if !store.validObjectKey(identity.Key) || identity.Version == "" {
		return artifactobject.DeleteResult{}, ErrObjectInput
	}
	output, err := store.api.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &store.bucket, Key: &identity.Key, VersionId: &identity.Version,
		ExpectedBucketOwner: store.owner})
	if err != nil {
		return store.recoverDelete(ctx, identity, err)
	}
	if output == nil {
		return store.recoverDelete(ctx, identity, ErrProviderEvidence)
	}
	requestID, err := metadataRequestID(output.ResultMetadata)
	if err != nil {
		return store.recoverDelete(ctx, identity, err)
	}
	returnedVersion := ""
	if output.VersionId != nil {
		returnedVersion = *output.VersionId
	}
	deleteMarker := output.DeleteMarker != nil && *output.DeleteMarker
	responseHash := evidenceHash("artifact-object-delete-response.v1", store.contract,
		store.bucket, identity.Key, identity.Version, requestID, returnedVersion,
		strconv.FormatBool(deleteMarker))
	absenceHash, verifiedAt, err := store.verifyAbsent(ctx, identity)
	if err != nil {
		return artifactobject.DeleteResult{}, err
	}
	return artifactobject.DeleteResult{ProviderRequestID: requestID,
		DeleteResponseSHA256: responseHash, AbsenceProbeSHA256: absenceHash,
		VerifiedAbsentAt: verifiedAt}, nil
}

func (store *Store) recoverDelete(ctx context.Context, identity artifactobject.Identity,
	effectErr error) (artifactobject.DeleteResult, error) {
	_, probeErr := store.api.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &store.bucket, Key: &identity.Key, VersionId: &identity.Version,
		ExpectedBucketOwner: store.owner})
	if probeErr == nil {
		return artifactobject.DeleteResult{}, providerFailure(effectErr)
	}
	if !provesAbsence(probeErr) {
		return artifactobject.DeleteResult{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, effectErr, probeErr)
	}
	absenceHash, verifiedAt, probeID, err := store.absenceEvidence(identity, probeErr)
	if err != nil {
		return artifactobject.DeleteResult{}, errors.Join(effectErr, err)
	}
	requestID := errorRequestID(effectErr)
	if requestID == "" {
		requestID = probeID
	}
	responseHash := evidenceHash("artifact-object-delete-recovery.v1", store.contract,
		store.bucket, identity.Key, identity.Version, requestID, probeID)
	return artifactobject.DeleteResult{ProviderRequestID: requestID,
		DeleteResponseSHA256: responseHash, AbsenceProbeSHA256: absenceHash,
		VerifiedAbsentAt: verifiedAt}, nil
}

func (store *Store) verifyAbsent(ctx context.Context,
	identity artifactobject.Identity) ([32]byte, time.Time, error) {
	_, err := store.api.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &store.bucket, Key: &identity.Key, VersionId: &identity.Version,
		ExpectedBucketOwner: store.owner})
	if err == nil || !provesAbsence(err) {
		if err == nil {
			err = errors.New("deleted artifact object version remains readable")
		}
		return [32]byte{}, time.Time{}, errors.Join(artifactobject.ErrOutcomeUnknown, err)
	}
	digest, verifiedAt, _, evidenceErr := store.absenceEvidence(identity, err)
	return digest, verifiedAt, evidenceErr
}

func (store *Store) absenceEvidence(identity artifactobject.Identity,
	probeErr error) ([32]byte, time.Time, string, error) {
	requestID := errorRequestID(probeErr)
	if requestID == "" {
		return [32]byte{}, time.Time{}, "", errors.Join(
			artifactobject.ErrOutcomeUnknown, ErrProviderEvidence)
	}
	verifiedAt, err := store.evidenceNow()
	if err != nil {
		return [32]byte{}, time.Time{}, "", err
	}
	digest := evidenceHash("artifact-object-absence-probe.v1", store.contract,
		store.bucket, identity.Key, identity.Version, requestID,
		strconv.Itoa(responseStatus(probeErr)))
	return digest, verifiedAt, requestID, nil
}
