package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"time"
)

const supportedArtifactContractRevision = 11

type sourceManifest struct {
	Version                  int          `json:"version"`
	ArtifactContractRevision int          `json:"artifact_contract_revision"`
	Profile                  string       `json:"profile,omitempty"`
	ProfileInputsSHA256      string       `json:"profile_inputs_sha256,omitempty"`
	LicenseKind              string       `json:"license_kind,omitempty"`
	RecoveryOperationID      string       `json:"recovery_operation_id,omitempty"`
	EnvironmentID            string       `json:"environment_id,omitempty"`
	DeploymentOperationID    string       `json:"deployment_operation_id,omitempty"`
	GenerationID             string       `json:"generation_id"`
	GeneratedAt              string       `json:"generated_at"`
	RenewAfter               string       `json:"renew_after"`
	ExpiresAt                string       `json:"expires_at"`
	Files                    []sourceFile `json:"files"`
}

type sourceFile struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

func decodeManifest(content []byte) (sourceManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var manifest sourceManifest
	if err := decoder.Decode(&manifest); err != nil {
		return sourceManifest{}, fmt.Errorf("decode source manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return sourceManifest{}, fmt.Errorf("source manifest has trailing content")
	}
	return manifest, nil
}

func validateManifest(
	manifest sourceManifest,
	now time.Time,
) error {
	identifier, err := base64.RawURLEncoding.DecodeString(manifest.GenerationID)
	if err != nil || len(identifier) != 16 || manifest.Version != 1 ||
		manifest.ArtifactContractRevision != supportedArtifactContractRevision {
		return fmt.Errorf("source manifest identity is invalid")
	}
	generated, err := canonicalTime(manifest.GeneratedAt)
	if err != nil {
		return err
	}
	renew, err := canonicalTime(manifest.RenewAfter)
	if err != nil {
		return err
	}
	expires, err := canonicalTime(manifest.ExpiresAt)
	if err != nil {
		return err
	}
	if generated.After(now.Add(5*time.Minute)) ||
		!validManifestLifetime(generated, renew, expires, now) {
		return fmt.Errorf("source manifest is expired or structurally invalid")
	}
	if len(manifest.Files) == 0 || len(manifest.Files) > 512 ||
		!sort.SliceIsSorted(manifest.Files, func(left, right int) bool {
			return manifest.Files[left].Name < manifest.Files[right].Name
		}) {
		return fmt.Errorf("source manifest catalog is invalid")
	}
	if manifest.Profile != "" && manifest.Profile != "local" ||
		manifest.ProfileInputsSHA256 != "" || manifest.LicenseKind != "" ||
		manifest.RecoveryOperationID != "" || manifest.EnvironmentID != "" ||
		manifest.DeploymentOperationID != "" {
		return fmt.Errorf("source manifest local profile differs")
	}
	for index, file := range manifest.Files {
		if file.Name == "manifest.json" || filepath.Base(file.Name) != file.Name ||
			file.Name == "." || file.Size < 1 || file.Size > 64<<20 ||
			len(file.SHA256) != 64 || (index > 0 && file.Name == manifest.Files[index-1].Name) {
			return fmt.Errorf("source manifest file entry is invalid")
		}
	}
	return nil
}

func validManifestLifetime(
	generated time.Time,
	renew time.Time,
	expires time.Time,
	now time.Time,
) bool {
	return renew.After(generated) && expires.Sub(renew) == 7*24*time.Hour &&
		now.Before(renew)
}

func canonicalTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.UTC().Format(time.RFC3339) != value {
		return time.Time{}, fmt.Errorf("source manifest timestamp is invalid")
	}
	return parsed, nil
}
