package s3object

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestNewVerifiedRequiresVersioningAndPrivateACL(t *testing.T) {
	fake := &fakeAPI{version: types.BucketVersioningStatusSuspended,
		acl: privateACL("123456789012")}
	_, err := NewVerified(context.Background(), fake, Config{Bucket: "exports",
		ObjectPrefix: "iam-system-exports", Contract: "IAM-SYSTEM-EXPORT-FCART1",
		MaximumCiphertextBytes: 1024}, fixedClock{})
	if !errors.Is(err, ErrVersioningRequired) {
		t.Fatalf("versioning error = %v", err)
	}
	fake.version = types.BucketVersioningStatusEnabled
	group := types.TypeGroup
	uri := "http://acs.amazonaws.com/groups/global/AllUsers"
	fake.acl.Grants = []types.Grant{{Grantee: &types.Grantee{Type: group, URI: &uri}}}
	_, err = NewVerified(context.Background(), fake, Config{Bucket: "exports",
		ObjectPrefix: "iam-system-exports", Contract: "IAM-SYSTEM-EXPORT-FCART1",
		MaximumCiphertextBytes: 1024}, fixedClock{})
	if !errors.Is(err, ErrPrivateACLRequired) {
		t.Fatalf("ACL error = %v", err)
	}
}

func TestPrivateBucketACLRejectsEveryForeignAuthority(t *testing.T) {
	owner := "123456789012"
	tests := []*s3.GetBucketAclOutput{nil, {}, {Owner: &types.Owner{ID: &owner},
		Grants: []types.Grant{{Grantee: nil}}}, privateACL(owner)}
	foreign := "210987654321"
	tests[3].Grants[0].Grantee.ID = &foreign
	for index, value := range tests {
		if privateBucketACL(value, owner) {
			t.Fatalf("foreign ACL %d accepted", index)
		}
	}
}

func TestNewVerifiedPinsBucketAndOwnerOnPreflight(t *testing.T) {
	fake := &fakeAPI{version: types.BucketVersioningStatusEnabled,
		acl: privateACL("123456789012")}
	_ = testStore(t, fake)
	if fake.versionInput.ExpectedBucketOwner == nil || fake.aclInput.ExpectedBucketOwner == nil ||
		*fake.versionInput.ExpectedBucketOwner != "123456789012" ||
		*fake.aclInput.ExpectedBucketOwner != "123456789012" {
		t.Fatal("preflight did not pin expected bucket owner")
	}
}

func TestNewVerifiedRejectsTypedNilDependencies(t *testing.T) {
	var client *fakeAPI
	_, err := NewVerified(context.Background(), client, Config{Bucket: "exports",
		ObjectPrefix: "iam-system-exports", Contract: "IAM-SYSTEM-EXPORT-FCART1",
		MaximumCiphertextBytes: 1024}, fixedClock{})
	if !errors.Is(err, ErrConfiguration) {
		t.Fatalf("typed nil error = %v", err)
	}
}
