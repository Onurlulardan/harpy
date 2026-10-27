//go:build integration

package s3blob

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type documentsIntegrationClock struct{}

func (documentsIntegrationClock) Now() time.Time { return time.Now().UTC() }

type documentsPrivateACLClient struct{ *s3.Client }

func (client documentsPrivateACLClient) GetBucketAcl(ctx context.Context,
	input *s3.GetBucketAclInput,
	options ...func(*s3.Options),
) (*s3.GetBucketAclOutput, error) {
	output, err := client.Client.GetBucketAcl(ctx, input, options...)
	if err != nil || output == nil || output.Owner == nil || len(output.Grants) != 1 ||
		output.Grants[0].Grantee == nil ||
		output.Grants[0].Grantee.Type != types.TypeCanonicalUser ||
		output.Grants[0].Grantee.URI != nil || output.Grants[0].Grantee.EmailAddress != nil {
		return output, err
	}
	owner := "minio-private-owner"
	output.Owner.ID = &owner
	output.Grants[0].Grantee.ID = &owner
	return output, nil
}

func TestMinIOExactDocumentVersionLifecycle(t *testing.T) {
	client := startDocumentsMinIO(t)
	bucket := "harpy-documents-integration"
	createDocumentsBucket(t, client, bucket)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	store, err := NewVerified(ctx, documentsPrivateACLClient{Client: client},
		Config{Bucket: bucket}, documentsIntegrationClock{})
	if err != nil {
		t.Fatal(err)
	}
	tenant := mustDocumentsID(t, "018f0000-0000-7000-8000-000000000101")
	item := mustDocumentsID(t, "018f0000-0000-7000-8000-000000000102")
	version := mustDocumentsID(t, "018f0000-0000-7000-8000-000000000103")
	deletionID := mustDocumentsID(t, "018f0000-0000-7000-8000-000000000104")
	key, err := domain.ContentObjectKey(tenant, item, version)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("FCDOC1 encrypted document integration payload")
	digest := sha256.Sum256(body)
	put, err := store.Put(ctx, uploadapp.ObjectPutRequest{Key: key, Generation: 1,
		CiphertextBytes: int64(len(body)), Body: bytes.NewReader(body)})
	if err != nil || put.Version == "" {
		t.Fatalf("Put() = %#v, %v", put, err)
	}
	reconciled, err := store.Reconcile(ctx, uploadapp.ObjectExpected{Key: key,
		Generation: 1, CiphertextBytes: int64(len(body)), CiphertextSHA256: digest})
	if err != nil || reconciled != put {
		t.Fatalf("Reconcile() = %#v, %v; want %#v", reconciled, err, put)
	}
	inventoried, found, err := store.FindExactVersion(ctx, key, 1, int64(len(body)))
	if err != nil || !found || inventoried != put.Version {
		t.Fatalf("FindExactVersion() = %q, %v, %v; want %q",
			inventoried, found, err, put.Version)
	}
	expected := downloadapp.ObjectExpected{Key: key, Version: put.Version,
		Generation: 1, CiphertextBytes: int64(len(body)), CiphertextSHA256: digest}
	reader, err := store.OpenExact(ctx, expected)
	if err != nil {
		t.Fatal(err)
	}
	read, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(read, body) {
		t.Fatalf("read = %q, %v, %v", read, readErr, closeErr)
	}
	receipt, err := store.DeleteExact(ctx, deletionapp.Task{ID: deletionID,
		TenantID: tenant, TargetID: version, TargetKind: "VERSION", ObjectKey: key,
		ObjectVersion: put.Version, ObjectGeneration: 1, CiphertextBytes: int64(len(body))})
	if err != nil || receipt.DeleteSHA256 == [32]byte{} || receipt.AbsenceSHA256 == [32]byte{} {
		t.Fatalf("DeleteExact() = %#v, %v", receipt, err)
	}
	_, err = store.OpenExact(ctx, expected)
	if !errors.Is(err, downloadapp.ErrUnavailable) {
		t.Fatalf("deleted OpenExact() error = %v", err)
	}
}

func mustDocumentsID(t *testing.T, value string) domain.ID {
	t.Helper()
	id, err := domain.ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
