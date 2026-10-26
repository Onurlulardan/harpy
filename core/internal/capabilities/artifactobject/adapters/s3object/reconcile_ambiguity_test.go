package s3object

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func TestReconcileExactFailsClosedOnVersionAmbiguity(t *testing.T) {
	modified := testNow.Add(-time.Minute)
	truncated, latest := true, true
	version, size := "version-1", int64(7)
	tests := map[string]*s3.ListObjectVersionsOutput{
		"truncated": {IsTruncated: &truncated, ResultMetadata: resultMetadata("list")},
		"delete marker": {IsTruncated: boolPointer(false),
			DeleteMarkers: []types.DeleteMarkerEntry{{Key: pointer(testObjectKey),
				VersionId: &version, IsLatest: &latest, LastModified: &modified}},
			ResultMetadata: resultMetadata("list")},
		"multiple versions": {IsTruncated: boolPointer(false),
			Versions: []types.ObjectVersion{
				{Key: pointer(testObjectKey), VersionId: &version, IsLatest: &latest,
					Size: &size, LastModified: &modified},
				{Key: pointer(testObjectKey), VersionId: pointer("version-2"),
					IsLatest: boolPointer(false), Size: &size, LastModified: &modified}},
			ResultMetadata: resultMetadata("list")},
	}
	for name, output := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &fakeAPI{list: func(*s3.ListObjectVersionsInput) (
				*s3.ListObjectVersionsOutput, error,
			) {
				return output, nil
			}, head: func(*s3.HeadObjectInput) (
				*s3.HeadObjectOutput, error,
			) {
				t.Fatal("ambiguous LIST reached HEAD")
				return nil, nil
			}}
			result, err := testStore(t, fake).ReconcileExact(t.Context(),
				artifactobject.ReconcileExpected{Key: testObjectKey,
					CiphertextBytes: 7, CiphertextSHA256: [32]byte{1},
					HeaderSHA256: [32]byte{2}})
			if result.State != artifactobject.ReconciliationUnknown ||
				!errors.Is(err, artifactobject.ErrOutcomeUnknown) {
				t.Fatalf("result = %#v, err = %v", result, err)
			}
		})
	}
}

func TestReconcileExactFailsClosedWhenObjectChangesBetweenListCalls(t *testing.T) {
	modified := testNow.Add(-time.Minute)
	expected := artifactobject.ReconcileExpected{Key: testObjectKey,
		CiphertextBytes: 7, CiphertextSHA256: [32]byte{1}, HeaderSHA256: [32]byte{2}}
	listCalls := 0
	fake := &fakeAPI{}
	fake.list = func(*s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error) {
		listCalls++
		if listCalls == 1 {
			return reconciliationListOutput(testObjectKey, "version-1", 7,
				modified, "list-1"), nil
		}
		return emptyReconciliationList("list-2"), nil
	}
	fake.head = func(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		store := &Store{contract: "IAM-SYSTEM-EXPORT-FCART1"}
		return &s3.HeadObjectOutput{VersionId: pointer("version-1"),
			ContentLength: &expected.CiphertextBytes, LastModified: &modified,
			Metadata: store.objectMetadata(expected.CiphertextSHA256,
				expected.HeaderSHA256, expected.CiphertextBytes),
			ChecksumSHA256: pointer(checksumBase64(expected.CiphertextSHA256)),
			ResultMetadata: resultMetadata("head")}, nil
	}
	result, err := testStore(t, fake).ReconcileExact(t.Context(), expected)
	if result.State != artifactobject.ReconciliationUnknown ||
		!errors.Is(err, artifactobject.ErrOutcomeUnknown) {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func boolPointer(value bool) *bool { return &value }
