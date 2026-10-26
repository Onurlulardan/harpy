package recoverystorage

import (
	"encoding/json"
	"testing"
)

func TestManifestRequiresIsolatedOffsiteObjectLockStorage(t *testing.T) {
	manifest := fixtureManifest()
	document, _ := json.Marshal(manifest)
	parsed, err := Parse(document)
	if err != nil || !parsed.Valid() {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	prefix, err := parsed.Prefix(
		"018f0000-0000-7000-8000-000000000001",
		"018f0000-0000-7000-8000-000000000002")
	if err != nil || prefix != "environments/018f0000-0000-7000-8000-000000000001/"+
		"recovery-sets/018f0000-0000-7000-8000-000000000002/" {
		t.Fatalf("unexpected prefix: %q %v", prefix, err)
	}
}

func TestManifestRejectsCrossModeAndCredentialReuse(t *testing.T) {
	tests := []func(*Manifest){
		func(value *Manifest) { value.DeploymentMode = "saas" },
		func(value *Manifest) { value.ObjectLock.Enabled = false },
		func(value *Manifest) { value.Bucket = value.DocumentsSource.Bucket },
		func(value *Manifest) {
			value.DocumentsRehydrateTarget.Bucket = value.DocumentsSource.Bucket
		},
		func(value *Manifest) { value.Credentials.RestoreRead = value.Credentials.RecoveryWrite },
		func(value *Manifest) { value.ProtectionClass = "offsite"; value.OffsiteProofSHA256 = "" },
	}
	for index, mutate := range tests {
		manifest := fixtureManifest()
		mutate(&manifest)
		document, _ := json.Marshal(manifest)
		if _, err := Parse(document); err == nil {
			t.Fatalf("invalid manifest %d accepted", index)
		}
	}
}

func TestLocalOnlyStorageCannotProduceProtectedStagingOrProductionSet(t *testing.T) {
	local := fixtureManifest()
	local.Endpoint = "http://minio:9000"
	local.ProtectionClass = "local_only"
	local.OffsiteProofSHA256 = ""
	if !local.Valid() || !local.AllowsProtected("development") {
		t.Fatal("development local-only rehearsal was rejected")
	}
	if local.AllowsProtected("staging") || local.AllowsProtected("production") {
		t.Fatal("local-only storage produced protected non-development evidence")
	}
	offsite := fixtureManifest()
	if !offsite.AllowsProtected("staging") || !offsite.AllowsProtected("production") {
		t.Fatal("valid offsite storage was rejected")
	}
}

func fixtureManifest() Manifest {
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return Manifest{Schema: Schema, DeploymentMode: "onpremise",
		Endpoint: "https://recovery.example.test", Region: "eu-central-1",
		Bucket: Bucket, ExpectedOwner: "owner", TLSCASHA256: digest,
		ProtectionClass: "offsite", OffsiteProofSHA256: digest,
		DocumentsSource: StorageProfile{Endpoint: "https://documents.example.test",
			Region: "eu-central-1", Bucket: "harpy-documents",
			ExpectedOwner: "owner", TLSCASHA256: digest},
		DocumentsRehydrateTarget: StorageProfile{
			Endpoint: "https://rehydrate.example.test", Region: "eu-central-1",
			Bucket: "harpy-documents-rehydrate", ExpectedOwner: "owner",
			TLSCASHA256: digest},
		ObjectLock: ObjectLockPolicy{Enabled: true, Mode: "COMPLIANCE", RetentionDays: 30},
		Credentials: CredentialFiles{RecoveryWrite: "/run/a", RetentionDelete: "/run/b",
			RestoreRead: "/run/c", DocumentsSourceRead: "/run/d", DocumentsRehydrate: "/run/e"},
		RecoveryKeyProvider: KeyProvider{Kind: "local_keyring", Handle: "recovery-keyring-v1"}}
}
