package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const (
	manifestVersion = 1
	renewalLead     = 7 * 24 * time.Hour
)

type setManifest struct {
	Version                  int            `json:"version"`
	ArtifactContractRevision int            `json:"artifact_contract_revision,omitempty"`
	Profile                  string         `json:"profile,omitempty"`
	ProfileInputsSHA256      string         `json:"profile_inputs_sha256,omitempty"`
	LicenseKind              string         `json:"license_kind,omitempty"`
	GenerationID             string         `json:"generation_id"`
	GeneratedAt              string         `json:"generated_at"`
	RenewAfter               string         `json:"renew_after"`
	ExpiresAt                string         `json:"expires_at"`
	Files                    []manifestFile `json:"files"`
}

type manifestFile struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

func buildSetManifest(set *artifactSet, now time.Time) (setManifest, error) {
	if err := set.requireComplete(); err != nil {
		return setManifest{}, err
	}
	renewAfter := set.expiresAt.Add(-renewalLead)
	if !renewAfter.After(now.Add(time.Hour)) {
		return setManifest{}, fmt.Errorf("artifact lifetime is too short")
	}
	generationID, err := randomBase64(16)
	if err != nil {
		return setManifest{}, err
	}
	names := set.expectedArtifactNames()
	files := make([]manifestFile, 0, len(names))
	for _, name := range names {
		content := set.files[name]
		files = append(files, manifestFile{
			Name: name, SHA256: sha256Hex(content), Size: len(content),
		})
	}
	return setManifest{
		Version:                  manifestVersion,
		ArtifactContractRevision: currentArtifactContractRevision,
		Profile:                  string(set.options.profile),
		ProfileInputsSHA256:      set.options.digest(),
		GenerationID:             generationID,
		GeneratedAt:              canonicalTimestamp(now),
		RenewAfter:               canonicalTimestamp(renewAfter),
		ExpiresAt:                canonicalTimestamp(set.expiresAt),
		Files:                    files,
	}, nil
}

func decodeSetManifest(content []byte) (setManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var manifest setManifest
	if err := decoder.Decode(&manifest); err != nil {
		return setManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return setManifest{}, fmt.Errorf("manifest has trailing content")
	}
	return manifest, nil
}

func validateManifestDocument(
	manifest setManifest,
	now time.Time,
	ignoreExpiry bool,
) error {
	if err := validateManifestLifecycle(manifest, now, ignoreExpiry); err != nil {
		return err
	}
	if err := validateCurrentArtifactContractRevision(manifest); err != nil {
		return err
	}
	if manifest.Profile != "" && manifest.Profile != string(profileLocal) ||
		manifest.ProfileInputsSHA256 != "" || manifest.LicenseKind != "" {
		return fmt.Errorf("manifest profile is not local")
	}
	return validateCurrentManifestCatalog(manifest.Files, profileLocal)
}

func validateResetManifestDocument(manifest setManifest, now time.Time) error {
	if err := validateManifestLifecycle(manifest, now, true); err != nil {
		return err
	}
	if err := validateResettableArtifactContractRevision(manifest); err != nil {
		return err
	}
	return validateManifestFileCatalog(manifest.Files)
}

func validateManifestLifecycle(
	manifest setManifest,
	now time.Time,
	ignoreExpiry bool,
) error {
	identifier, err := base64.RawURLEncoding.DecodeString(manifest.GenerationID)
	if err != nil || len(identifier) != 16 ||
		base64.RawURLEncoding.EncodeToString(identifier) != manifest.GenerationID {
		return fmt.Errorf("manifest generation ID is invalid")
	}
	generatedAt, err := parseCanonicalTimestamp(manifest.GeneratedAt)
	if err != nil {
		return err
	}
	renewAfter, err := parseCanonicalTimestamp(manifest.RenewAfter)
	if err != nil {
		return err
	}
	expiresAt, err := parseCanonicalTimestamp(manifest.ExpiresAt)
	if err != nil {
		return err
	}
	if manifest.Version != manifestVersion || generatedAt.After(now.Add(5*time.Minute)) ||
		!renewAfter.After(generatedAt) || expiresAt.Sub(renewAfter) != renewalLead ||
		(!ignoreExpiry && !now.Before(renewAfter)) {
		return fmt.Errorf("manifest is expired or structurally invalid")
	}
	return nil
}

func parseCanonicalTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || canonicalTimestamp(parsed) != value {
		return time.Time{}, fmt.Errorf("manifest timestamp is invalid")
	}
	return parsed, nil
}
