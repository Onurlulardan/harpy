package recoverystorage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const (
	Schema               = "recovery-storage-manifest.v1"
	Bucket               = "harpy-recovery"
	MaximumManifestBytes = 64 * 1024
)

var ErrManifest = errors.New("recovery storage manifest is invalid")

type ObjectLockPolicy struct {
	Enabled       bool   `json:"enabled"`
	Mode          string `json:"mode"`
	RetentionDays int    `json:"retention_days"`
}

type CredentialFiles struct {
	RecoveryWrite       string `json:"recovery_write"`
	RetentionDelete     string `json:"retention_delete"`
	RestoreRead         string `json:"restore_read"`
	DocumentsSourceRead string `json:"documents_source_read"`
	DocumentsRehydrate  string `json:"documents_rehydrate_write"`
}

type KeyProvider struct {
	Kind   string `json:"kind"`
	Handle string `json:"handle"`
}

type StorageProfile struct {
	Endpoint      string `json:"endpoint"`
	Region        string `json:"region"`
	Bucket        string `json:"bucket"`
	ExpectedOwner string `json:"expected_owner"`
	PathStyle     bool   `json:"path_style"`
	TLSCASHA256   string `json:"tls_ca_sha256"`
}

type Manifest struct {
	Schema                   string           `json:"schema"`
	DeploymentMode           string           `json:"deployment_mode"`
	Endpoint                 string           `json:"endpoint"`
	Region                   string           `json:"region"`
	Bucket                   string           `json:"bucket"`
	ExpectedOwner            string           `json:"expected_owner"`
	PathStyle                bool             `json:"path_style"`
	TLSCASHA256              string           `json:"tls_ca_sha256"`
	ProtectionClass          string           `json:"protection_class"`
	OffsiteProofSHA256       string           `json:"offsite_proof_sha256,omitempty"`
	DocumentsSource          StorageProfile   `json:"documents_source"`
	DocumentsRehydrateTarget StorageProfile   `json:"documents_rehydrate_target"`
	ObjectLock               ObjectLockPolicy `json:"object_lock"`
	Credentials              CredentialFiles  `json:"credentials"`
	RecoveryKeyProvider      KeyProvider      `json:"recovery_key_provider"`
}

func Parse(document []byte) (Manifest, error) {
	if len(document) == 0 || len(document) > MaximumManifestBytes {
		return Manifest{}, ErrManifest
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if decoder.Decode(&manifest) != nil || decoder.Decode(new(any)) != io.EOF ||
		!manifest.Valid() {
		return Manifest{}, ErrManifest
	}
	return manifest, nil
}

func (manifest Manifest) Prefix(environmentID, recoverySetID string) (string, error) {
	if !validUUID(environmentID) || !validUUID(recoverySetID) {
		return "", ErrManifest
	}
	return "environments/" + environmentID + "/recovery-sets/" + recoverySetID + "/", nil
}
