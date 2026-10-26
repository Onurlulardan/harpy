package s3object

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func uploadFixture() (artifactobject.Upload, []byte) {
	body := []byte("encrypted-fcart-body")
	return artifactobject.Upload{Key: testObjectKey, Body: bytes.NewReader(body),
		CiphertextBytes: int64(len(body)), CiphertextSHA256: sha256.Sum256(body),
		HeaderSHA256: sha256.Sum256([]byte("header"))}, body
}

func TestUploadCreatesImmutableChecksummedVersion(t *testing.T) {
	request, body := uploadFixture()
	version, checksum := "version-1", checksumBase64(request.CiphertextSHA256)
	fake := &fakeAPI{put: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
		actual, err := io.ReadAll(input.Body)
		if err != nil || !bytes.Equal(actual, body) || input.IfNoneMatch == nil ||
			*input.IfNoneMatch != "*" || input.ChecksumSHA256 == nil ||
			*input.ChecksumSHA256 != checksum ||
			input.Metadata[metadataContract] != "IAM-SYSTEM-EXPORT-FCART1" {
			t.Fatal("PutObject did not preserve the immutable object contract")
		}
		return &s3.PutObjectOutput{VersionId: &version, ChecksumSHA256: &checksum,
			ResultMetadata: resultMetadata("put-1")}, nil
	}}
	result, err := testStore(t, fake).Upload(context.Background(), request)
	if err != nil || result.Identity.Version != version ||
		result.ProviderRequestID != "put-1" || result.UploadedAt != testNow {
		t.Fatalf("Upload() = %#v, %v", result, err)
	}
}

func TestUploadRecoversOnlyExactLostSuccessEvidence(t *testing.T) {
	request, _ := uploadFixture()
	version := "version-1"
	fake := &fakeAPI{put: func(*s3.PutObjectInput) (*s3.PutObjectOutput, error) {
		return nil, responseError(http.StatusInternalServerError, "put-lost")
	}, head: func(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		return matchingHead(request, version), nil
	}}
	result, err := testStore(t, fake).Upload(context.Background(), request)
	if err != nil || result.Identity.Version != version || result.ProviderRequestID != "put-lost" {
		t.Fatalf("recovered Upload() = %#v, %v", result, err)
	}
	fake.head = func(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		probe := matchingHead(request, "foreign")
		probe.Metadata[metadataHeader] = "wrong"
		return probe, nil
	}
	if _, err := testStore(t, fake).Upload(context.Background(), request); !errors.Is(err, artifactobject.ErrOutcomeUnknown) {
		t.Fatalf("mismatched recovery error = %v", err)
	}
}

func matchingHead(request artifactobject.Upload, version string) *s3.HeadObjectOutput {
	return &s3.HeadObjectOutput{VersionId: &version,
		ContentLength:  &request.CiphertextBytes,
		ChecksumSHA256: pointer(checksumBase64(request.CiphertextSHA256)),
		Metadata: testMetadata(request.CiphertextSHA256, request.HeaderSHA256,
			request.CiphertextBytes), ResultMetadata: resultMetadata("head-1")}
}

func testMetadata(ciphertext, header [32]byte, size int64) map[string]string {
	store := &Store{contract: "IAM-SYSTEM-EXPORT-FCART1"}
	return store.objectMetadata(ciphertext, header, size)
}

type invalidClock struct{ value time.Time }

func (clock invalidClock) Now() time.Time { return clock.value }

func TestUploadRejectsInvalidEvidenceTimeAfterProviderEffect(t *testing.T) {
	request, _ := uploadFixture()
	version, checksum := "version-1", checksumBase64(request.CiphertextSHA256)
	fake := &fakeAPI{put: func(*s3.PutObjectInput) (*s3.PutObjectOutput, error) {
		return &s3.PutObjectOutput{VersionId: &version, ChecksumSHA256: &checksum,
			ResultMetadata: resultMetadata("put-1")}, nil
	}}
	store := testStore(t, fake)
	store.clock = invalidClock{value: time.Time{}}
	if _, err := store.Upload(context.Background(), request); !errors.Is(err, artifactobject.ErrOutcomeUnknown) {
		t.Fatalf("invalid clock error = %v", err)
	}
}
