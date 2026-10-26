package s3object

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func expectedFixture(body []byte) artifactobject.Expected {
	return artifactobject.Expected{Identity: artifactobject.Identity{
		Key: testObjectKey, Version: "version-1"}, CiphertextBytes: int64(len(body)),
		CiphertextSHA256: sha256.Sum256(body), HeaderSHA256: sha256.Sum256([]byte("header"))}
}

func TestOpenExactPinsVersionAndIntegrityEvidence(t *testing.T) {
	body := []byte("encrypted-fcart-body")
	expected := expectedFixture(body)
	fake := &fakeAPI{get: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		if input.VersionId == nil || *input.VersionId != expected.Identity.Version {
			t.Fatalf("GetObject version = %v", input.VersionId)
		}
		return matchingGet(expected, io.NopCloser(bytes.NewReader(body))), nil
	}}
	reader, err := testStore(t, fake).OpenExact(context.Background(), expected)
	if err != nil {
		t.Fatal(err)
	}
	actual, readErr := io.ReadAll(reader)
	if closeErr := reader.Close(); readErr != nil || closeErr != nil || !bytes.Equal(actual, body) {
		t.Fatalf("read = %q, %v, %v", actual, readErr, closeErr)
	}
	fake.get = func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return nil, responseError(http.StatusNotFound, "get-1")
	}
	if _, err := testStore(t, fake).OpenExact(context.Background(), expected); !errors.Is(err, artifactobject.ErrIntegrity) {
		t.Fatalf("missing exact object error = %v", err)
	}
}

func matchingGet(expected artifactobject.Expected, body io.ReadCloser) *s3.GetObjectOutput {
	return &s3.GetObjectOutput{Body: body, VersionId: &expected.Identity.Version,
		ContentLength:  &expected.CiphertextBytes,
		ChecksumSHA256: pointer(checksumBase64(expected.CiphertextSHA256)),
		Metadata: testMetadata(expected.CiphertextSHA256, expected.HeaderSHA256,
			expected.CiphertextBytes)}
}

func TestDeleteExactRecoversLostResponseFromAbsence(t *testing.T) {
	identity := artifactobject.Identity{Key: testObjectKey, Version: "version-1"}
	fake := &fakeAPI{delete: func(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
		return nil, responseError(http.StatusInternalServerError, "delete-lost")
	}, head: func(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		return nil, responseError(http.StatusNotFound, "absence-1")
	}}
	result, err := testStore(t, fake).DeleteExact(context.Background(), identity)
	if err != nil || result.ProviderRequestID != "delete-lost" ||
		result.DeleteResponseSHA256 == [32]byte{} || result.AbsenceProbeSHA256 == [32]byte{} {
		t.Fatalf("DeleteExact() = %#v, %v", result, err)
	}
}

func TestDeleteExactDoesNotInventAbsence(t *testing.T) {
	identity := artifactobject.Identity{Key: testObjectKey, Version: "version-1"}
	fake := &fakeAPI{delete: func(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
		return nil, errors.New("connection reset")
	}, head: func(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		return nil, responseError(http.StatusForbidden, "probe-1")
	}}
	if _, err := testStore(t, fake).DeleteExact(context.Background(), identity); !errors.Is(err, artifactobject.ErrOutcomeUnknown) {
		t.Fatalf("ambiguous delete error = %v", err)
	}
	fake.head = func(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		return &s3.HeadObjectOutput{}, nil
	}
	if _, err := testStore(t, fake).DeleteExact(context.Background(), identity); !errors.Is(err, artifactobject.ErrUnavailable) {
		t.Fatalf("proven remaining delete error = %v", err)
	}
}
