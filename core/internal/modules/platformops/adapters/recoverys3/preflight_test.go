package recoverys3

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverystorage"
)

type fakeS3 struct {
	version types.BucketVersioningStatus
	lock    types.ObjectLockEnabled
	mode    types.ObjectLockRetentionMode
	days    int32
}

func (fake fakeS3) GetBucketAcl(context.Context, *s3.GetBucketAclInput,
	...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	owner := "owner"
	return &s3.GetBucketAclOutput{Owner: &types.Owner{ID: &owner},
		Grants: []types.Grant{{Grantee: &types.Grantee{
			Type: types.TypeCanonicalUser, ID: &owner}, Permission: types.PermissionFullControl}}}, nil
}
func (fake fakeS3) GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput,
	...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return &s3.GetBucketVersioningOutput{Status: fake.version}, nil
}
func (fake fakeS3) GetObjectLockConfiguration(context.Context,
	*s3.GetObjectLockConfigurationInput, ...func(*s3.Options)) (*s3.GetObjectLockConfigurationOutput, error) {
	return &s3.GetObjectLockConfigurationOutput{ObjectLockConfiguration: &types.ObjectLockConfiguration{ObjectLockEnabled: fake.lock,
		Rule: &types.ObjectLockRule{DefaultRetention: &types.DefaultRetention{
			Mode: fake.mode, Days: &fake.days}}}}, nil
}

func TestPreflightRequiresPrivateVersionedComplianceBucket(t *testing.T) {
	manifest := storageManifest()
	client := fakeS3{version: types.BucketVersioningStatusEnabled,
		lock: types.ObjectLockEnabledEnabled,
		mode: types.ObjectLockRetentionModeCompliance, days: 30}
	if _, err := Preflight(context.Background(), client, manifest); err != nil {
		t.Fatal(err)
	}
	client.version = types.BucketVersioningStatusSuspended
	if _, err := Preflight(context.Background(), client, manifest); err == nil {
		t.Fatal("suspended versioning accepted")
	}
}

func storageManifest() recoverystorage.Manifest {
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return recoverystorage.Manifest{Schema: recoverystorage.Schema, DeploymentMode: "onpremise",
		Endpoint: "https://recovery.example.test", Region: "eu-central-1",
		Bucket: recoverystorage.Bucket, ExpectedOwner: "owner", TLSCASHA256: digest,
		ProtectionClass: "offsite", OffsiteProofSHA256: digest,
		DocumentsSource: recoverystorage.StorageProfile{
			Endpoint: "https://documents.example.test", Region: "eu-central-1",
			Bucket: "documents", ExpectedOwner: "owner", TLSCASHA256: digest},
		DocumentsRehydrateTarget: recoverystorage.StorageProfile{
			Endpoint: "https://rehydrate.example.test", Region: "eu-central-1",
			Bucket: "documents-rehydrate", ExpectedOwner: "owner", TLSCASHA256: digest},
		ObjectLock: recoverystorage.ObjectLockPolicy{Enabled: true, Mode: "COMPLIANCE", RetentionDays: 30},
		Credentials: recoverystorage.CredentialFiles{RecoveryWrite: "/run/a",
			RetentionDelete: "/run/b", RestoreRead: "/run/c",
			DocumentsSourceRead: "/run/d", DocumentsRehydrate: "/run/e"},
		RecoveryKeyProvider: recoverystorage.KeyProvider{Kind: "local_keyring", Handle: "keyring"}}
}
