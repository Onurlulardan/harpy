package s3object

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithymiddleware "github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const testObjectKey = "iam-system-exports/v1/01910000-0000-7000-8000-000000000001/1.fcart"

var testNow = time.Date(2026, 7, 18, 3, 4, 5, 0, time.UTC)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return testNow }

type fakeAPI struct {
	put          func(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	head         func(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
	get          func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	delete       func(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	list         func(*s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error)
	version      types.BucketVersioningStatus
	acl          *s3.GetBucketAclOutput
	versionInput *s3.GetBucketVersioningInput
	aclInput     *s3.GetBucketAclInput
}

func (fake *fakeAPI) PutObject(_ context.Context, input *s3.PutObjectInput,
	_ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return fake.put(input)
}

func (fake *fakeAPI) HeadObject(_ context.Context, input *s3.HeadObjectInput,
	_ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return fake.head(input)
}

func (fake *fakeAPI) GetObject(_ context.Context, input *s3.GetObjectInput,
	_ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return fake.get(input)
}

func (fake *fakeAPI) DeleteObject(_ context.Context, input *s3.DeleteObjectInput,
	_ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return fake.delete(input)
}

func (fake *fakeAPI) ListObjectVersions(
	_ context.Context,
	input *s3.ListObjectVersionsInput,
	_ ...func(*s3.Options),
) (*s3.ListObjectVersionsOutput, error) {
	return fake.list(input)
}

func (fake *fakeAPI) GetBucketVersioning(_ context.Context,
	input *s3.GetBucketVersioningInput,
	_ ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	fake.versionInput = input
	return &s3.GetBucketVersioningOutput{Status: fake.version}, nil
}

func (fake *fakeAPI) GetBucketAcl(_ context.Context, input *s3.GetBucketAclInput,
	_ ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	fake.aclInput = input
	return fake.acl, nil
}

func testStore(t *testing.T, fake *fakeAPI) *Store {
	t.Helper()
	fake.version = types.BucketVersioningStatusEnabled
	if fake.acl == nil {
		fake.acl = privateACL("123456789012")
	}
	store, err := NewVerified(context.Background(), fake, Config{Bucket: "exports",
		ExpectedBucketOwner: "123456789012", ObjectPrefix: "iam-system-exports",
		Contract: "IAM-SYSTEM-EXPORT-FCART1", MaximumCiphertextBytes: 1 << 30}, fixedClock{})
	if err != nil {
		t.Fatalf("NewVerified() error = %v", err)
	}
	return store
}

func privateACL(owner string) *s3.GetBucketAclOutput {
	permission := types.PermissionFullControl
	typeCanonical := types.TypeCanonicalUser
	return &s3.GetBucketAclOutput{Owner: &types.Owner{ID: &owner},
		Grants: []types.Grant{{Permission: permission,
			Grantee: &types.Grantee{ID: &owner, Type: typeCanonical}}}}
}

func resultMetadata(requestID string) smithymiddleware.Metadata {
	var metadata smithymiddleware.Metadata
	awsmiddleware.SetRequestIDMetadata(&metadata, requestID)
	return metadata
}

func responseError(status int, requestID string) error {
	return &smithyhttp.ResponseError{Response: &smithyhttp.Response{Response: &http.Response{
		StatusCode: status, Header: http.Header{"X-Amz-Request-Id": []string{requestID}}}},
		Err: errors.New("provider response")}
}

func pointer(value string) *string { return &value }
