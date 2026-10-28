package main

import (
	"encoding/json"
	"strings"
	"testing"
)

type recoveryCredentialFixture struct {
	Schema          string `json:"schema"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

func TestLocalRecoveryCredentialsArePurposeSeparated(t *testing.T) {
	set, _ := generatedSetForTest(t)
	names := []string{
		"recovery-write-credential.json",
		"recovery-retention-delete-credential.json",
		"recovery-restore-read-credential.json",
		"recovery-documents-source-read-credential.json",
		"recovery-documents-rehydrate-write-credential.json",
	}
	seen := map[string]string{}
	for _, name := range names {
		var value recoveryCredentialFixture
		if json.Unmarshal(set.files[name], &value) != nil ||
			value.Schema != "recovery-s3-credential.v1" ||
			value.AccessKeyID == "" || value.SecretAccessKey == "" {
			t.Fatalf("%s is not a recovery credential", name)
		}
		for _, secret := range []string{value.AccessKeyID, value.SecretAccessKey} {
			if prior := seen[secret]; prior != "" {
				t.Fatalf("%s reuses material from %s", name, prior)
			}
			seen[secret] = name
		}
	}
	for _, name := range []string{
		"s3-access-key-id.txt", "s3-secret-access-key.txt",
		"minio-root-access-key.txt", "minio-root-secret-key.txt",
		"recovery-minio-root-access-key.txt", "recovery-minio-root-secret-key.txt",
	} {
		value := string(set.files[name])
		if value == "" || seen[value] != "" {
			t.Fatalf("%s collides with a purpose credential", name)
		}
		seen[value] = name
	}
}

func TestLocalRecoveryManifestBindsIsolatedStorage(t *testing.T) {
	set, _ := generatedSetForTest(t)
	var value struct {
		Schema          string `json:"schema"`
		DeploymentMode  string `json:"deployment_mode"`
		Endpoint        string `json:"endpoint"`
		Bucket          string `json:"bucket"`
		ExpectedOwner   string `json:"expected_owner"`
		ProtectionClass string `json:"protection_class"`
		ObjectLock      struct {
			Enabled       bool   `json:"enabled"`
			Mode          string `json:"mode"`
			RetentionDays int    `json:"retention_days"`
		} `json:"object_lock"`
		DocumentsSource struct {
			Endpoint      string `json:"endpoint"`
			Bucket        string `json:"bucket"`
			ExpectedOwner string `json:"expected_owner"`
		} `json:"documents_source"`
		DocumentsTarget struct {
			Endpoint      string `json:"endpoint"`
			Bucket        string `json:"bucket"`
			ExpectedOwner string `json:"expected_owner"`
		} `json:"documents_rehydrate_target"`
		Credentials map[string]string `json:"credentials"`
	}
	if err := json.Unmarshal(set.files["recovery-storage-manifest.v1.json"], &value); err != nil {
		t.Fatal(err)
	}
	if value.Schema != "recovery-storage-manifest.v1" ||
		value.DeploymentMode != "onpremise" || value.Endpoint != "http://recovery-s3-owner:9000" ||
		value.Bucket != "harpy-recovery" || value.ExpectedOwner != "210987654321" ||
		value.ProtectionClass != "local_only" || !value.ObjectLock.Enabled ||
		value.ObjectLock.Mode != "COMPLIANCE" || value.ObjectLock.RetentionDays != 30 {
		t.Fatalf("recovery storage contract drifted: %#v", value)
	}
	if value.DocumentsSource.Endpoint != "http://s3-owner:9000" ||
		value.DocumentsSource.Bucket != "harpy-documents" ||
		value.DocumentsTarget.Bucket != "harpy-documents-rehydrate" ||
		value.DocumentsSource.ExpectedOwner != "123456789012" ||
		value.DocumentsTarget.ExpectedOwner != "123456789012" ||
		len(value.Credentials) != 5 {
		t.Fatalf("Documents recovery storage contract drifted: %#v", value)
	}
}

func TestRecoveryProfilesUseExactBranchesAndReconcileEvidence(t *testing.T) {
	set, _ := generatedSetForTest(t)
	restore := string(set.files["database-restore.env"])
	rehydrate := string(set.files["database-rehydrate.env"])
	if !strings.Contains(restore, "HARPY_RESTORE_PLAN_PATH=/run/handoff/restore-plan.v1.json\n") ||
		strings.Contains(restore, "HARPY_REHYDRATE_REQUEST_PATH") ||
		!strings.Contains(rehydrate, "HARPY_REHYDRATE_REQUEST_PATH=/run/handoff/rehydrate-request.v1.json\n") ||
		strings.Contains(rehydrate, "HARPY_RESTORE_PLAN_PATH") {
		t.Fatal("restore and rehydrate profiles are not exact XOR branches")
	}
	identity := string(set.files["operator-restore-identity.env"])
	for _, value := range []string{
		"HARPY_RESTORE_RECEIPT_PATH=/run/restore-state/receipt.json",
		"HARPY_RESTORE_SECURITY_EVIDENCE_PATH=/run/restore-state/security-evidence.json",
		"HARPY_IAM_KEY_CUSTODY_MANIFEST_PATH=/run/restore-state/iam-key-custody-manifest.json",
		"HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH=/run/restore-state/protected-authority-ack-trust-bundle.v1.json",
	} {
		if !strings.Contains(identity, value+"\n") {
			t.Fatalf("identity reconcile profile missing %s", value)
		}
	}
}
