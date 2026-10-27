//go:build integration

package s3object

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
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

type integrationClock struct{}

func (integrationClock) Now() time.Time { return time.Now().UTC() }

type minIOPrivateACLClient struct{ *s3.Client }

func (client minIOPrivateACLClient) GetBucketAcl(ctx context.Context,
	input *s3.GetBucketAclInput,
	options ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
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

func TestMinIOExactVersionLifecycle(t *testing.T) {
	client := startMinIO(t)
	bucket := "forgecrud-artifact-object-integration"
	createVersionedBucket(t, client, bucket)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	store, err := NewVerified(ctx, minIOPrivateACLClient{Client: client}, Config{Bucket: bucket,
		ObjectPrefix: "iam-system-exports", Contract: "IAM-SYSTEM-EXPORT-FCART1",
		MaximumCiphertextBytes: 1 << 30}, integrationClock{})
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("FCART1 encrypted integration payload")
	ciphertext, header := sha256.Sum256(body), sha256.Sum256([]byte("header"))
	upload, err := store.Upload(ctx, artifactobject.Upload{Key: testObjectKey,
		Body: bytes.NewReader(body), CiphertextBytes: int64(len(body)),
		CiphertextSHA256: ciphertext, HeaderSHA256: header})
	if err != nil || upload.Identity.Version == "" {
		t.Fatalf("Upload() = %#v, %v", upload, err)
	}
	reconciled, err := store.ReconcileExact(ctx, artifactobject.ReconcileExpected{
		Key: testObjectKey, CiphertextBytes: int64(len(body)),
		CiphertextSHA256: ciphertext, HeaderSHA256: header})
	if err != nil || reconciled.State != artifactobject.ReconciliationSealed ||
		reconciled.Identity != upload.Identity || reconciled.EvidenceSHA256 == [32]byte{} {
		t.Fatalf("ReconcileExact() = %#v, %v", reconciled, err)
	}
	reader, err := store.OpenExact(ctx, artifactobject.Expected{Identity: upload.Identity,
		CiphertextBytes: int64(len(body)), CiphertextSHA256: ciphertext,
		HeaderSHA256: header})
	if err != nil {
		t.Fatal(err)
	}
	read, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(read, body) {
		t.Fatalf("read = %q, %v, %v", read, readErr, closeErr)
	}
	deleted, err := store.DeleteExact(ctx, upload.Identity)
	if err != nil || deleted.DeleteResponseSHA256 == [32]byte{} ||
		deleted.AbsenceProbeSHA256 == [32]byte{} {
		t.Fatalf("DeleteExact() = %#v, %v", deleted, err)
	}
	_, err = store.OpenExact(ctx, artifactobject.Expected{Identity: upload.Identity,
		CiphertextBytes: int64(len(body)), CiphertextSHA256: ciphertext,
		HeaderSHA256: header})
	if !errors.Is(err, artifactobject.ErrIntegrity) {
		t.Fatalf("deleted OpenExact() error = %v", err)
	}
}
