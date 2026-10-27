package s3object

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func TestReconcileExactProvesOneImmutableSealedVersion(t *testing.T) {
	modified := testNow.Add(-time.Minute)
	expected := artifactobject.ReconcileExpected{Key: testObjectKey,
		CiphertextBytes: 7, CiphertextSHA256: [32]byte{1}, HeaderSHA256: [32]byte{2}}
	listCalls := 0
	fake := &fakeAPI{}
	fake.list = func(input *s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error) {
		listCalls++
		if input.Prefix == nil || *input.Prefix != expected.Key ||
			input.MaxKeys == nil || *input.MaxKeys != reconciliationListLimit {
			t.Fatalf("list input = %#v", input)
		}
		return reconciliationListOutput(expected.Key, "version-1", expected.CiphertextBytes,
			modified, "list-request"), nil
	}
	fake.head = func(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		if input.VersionId == nil || *input.VersionId != "version-1" {
			t.Fatalf("head version = %v", input.VersionId)
		}
		store := &Store{contract: "IAM-SYSTEM-EXPORT-FCART1"}
		return &s3.HeadObjectOutput{VersionId: pointer("version-1"),
			ContentLength: &expected.CiphertextBytes, LastModified: &modified,
			Metadata: store.objectMetadata(expected.CiphertextSHA256,
				expected.HeaderSHA256, expected.CiphertextBytes),
			ChecksumSHA256: pointer(checksumBase64(expected.CiphertextSHA256)),
			ResultMetadata: resultMetadata("head-request")}, nil
	}
	result, err := testStore(t, fake).ReconcileExact(t.Context(), expected)
	if err != nil || result.State != artifactobject.ReconciliationSealed ||
		result.Identity.Version != "version-1" || result.ProviderRequestID != "head-request" ||
		result.EvidenceSHA256 == [32]byte{} || result.ObjectModifiedAt != modified ||
		result.VerifiedAt != testNow || listCalls != 2 {
		t.Fatalf("result = %#v, err = %v, list calls = %d", result, err, listCalls)
	}
}

func TestReconcileExactProvesMissingWithThreeBoundedObservations(t *testing.T) {
	listCalls := 0
	fake := &fakeAPI{}
	fake.list = func(*s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error) {
		listCalls++
		return emptyReconciliationList("list-missing"), nil
	}
	fake.head = func(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		if input.VersionId != nil {
			t.Fatalf("missing HEAD pinned version = %q", *input.VersionId)
		}
		return nil, responseError(404, "head-missing")
	}
	expected := artifactobject.ReconcileExpected{Key: testObjectKey,
		CiphertextBytes: 7, CiphertextSHA256: [32]byte{1}, HeaderSHA256: [32]byte{2}}
	result, err := testStore(t, fake).ReconcileExact(t.Context(), expected)
	if err != nil || result.State != artifactobject.ReconciliationMissing ||
		result.ProviderRequestID != "head-missing" || result.EvidenceSHA256 == [32]byte{} ||
		result.VerifiedAt != testNow || listCalls != 2 {
		t.Fatalf("result = %#v, err = %v, list calls = %d", result, err, listCalls)
	}
}

func reconciliationListOutput(key, version string, size int64, modified time.Time,
	requestID string,
) *s3.ListObjectVersionsOutput {
	truncated, latest := false, true
	return &s3.ListObjectVersionsOutput{IsTruncated: &truncated,
		Versions: []types.ObjectVersion{{Key: &key, VersionId: &version,
			IsLatest: &latest, Size: &size, LastModified: &modified}},
		ResultMetadata: resultMetadata(requestID)}
}

func emptyReconciliationList(requestID string) *s3.ListObjectVersionsOutput {
	truncated := false
	return &s3.ListObjectVersionsOutput{IsTruncated: &truncated,
		ResultMetadata: resultMetadata(requestID)}
}
