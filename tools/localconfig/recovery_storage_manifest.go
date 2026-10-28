package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	localRecoveryOwner = "210987654321"
	recoveryBucket     = "harpy-recovery"
	rehydrateBucket    = "harpy-documents-rehydrate"
	recoveryRetention  = 30
)

func addRecoveryStorageManifest(set *artifactSet) error {
	certificate, exists := set.files["local-ca.crt"]
	if !exists {
		return fmt.Errorf("local CA is missing before recovery manifest generation")
	}
	digest := sha256.Sum256(certificate)
	caSHA256 := hex.EncodeToString(digest[:])
	document := map[string]any{
		"schema":           "recovery-storage-manifest.v1",
		"deployment_mode":  "onpremise",
		"endpoint":         "http://recovery-s3-owner:9000",
		"region":           "us-east-1",
		"bucket":           recoveryBucket,
		"expected_owner":   localRecoveryOwner,
		"path_style":       true,
		"tls_ca_sha256":    caSHA256,
		"protection_class": "local_only",
		"documents_source": storageProfile(
			"http://s3-owner:9000", "harpy-documents", localOwner, caSHA256),
		"documents_rehydrate_target": storageProfile(
			"http://s3-owner:9000", rehydrateBucket, localOwner, caSHA256),
		"object_lock": map[string]any{
			"enabled": true, "mode": "COMPLIANCE",
			"retention_days": recoveryRetention,
		},
		"credentials": map[string]string{
			"recovery_write": "/run/secrets/recovery-write-credential.json",
			"retention_delete": "/run/secrets/" +
				"recovery-retention-delete-credential.json",
			"restore_read": "/run/secrets/recovery-restore-read-credential.json",
			"documents_source_read": "/run/secrets/" +
				"recovery-documents-source-read-credential.json",
			"documents_rehydrate_write": "/run/secrets/" +
				"recovery-documents-rehydrate-write-credential.json",
		},
		"recovery_key_provider": map[string]string{
			"kind": "local_keyring", "handle": "local-recovery-keyring-v1",
		},
	}
	return addJSON(set, "recovery-storage-manifest.v1.json", document)
}

func storageProfile(endpoint, bucket, owner, caSHA256 string) map[string]any {
	return map[string]any{
		"endpoint": endpoint, "region": "us-east-1", "bucket": bucket,
		"expected_owner": owner, "path_style": true, "tls_ca_sha256": caSHA256,
	}
}
