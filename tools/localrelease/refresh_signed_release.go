package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const signedReleaseValidity = 24 * time.Hour

func runRefreshSignedRelease(arguments []string) error {
	directory, releaseID, err := parseRefreshSignedReleaseOptions(arguments)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Truncate(time.Second)
	if err := refreshSignedRelease(directory, releaseID, now); err != nil {
		return err
	}
	fmt.Printf("local signed release refreshed: %s\n", directory)
	return nil
}

func refreshSignedRelease(directory, releaseID string, now time.Time) error {
	if err := verifyOwnedDirectory(directory, filepath.Dir(directory)); err != nil {
		return err
	}
	manifest, err := loadReleaseManifest(directory)
	if err != nil {
		return err
	}
	if id := strings.TrimSpace(releaseID); id != "" {
		manifest.ReleaseID = id
	}
	issuedAt := now.Add(-time.Minute)
	manifest.IssuedAt = issuedAt.Format(time.RFC3339)
	manifest.ExpiresAt = now.Add(signedReleaseValidity).Format(time.RFC3339)
	signer, err := newSigningAuthority(now)
	if err != nil {
		return err
	}
	release, err := newArtifact("release.json", manifest)
	if err != nil {
		return err
	}
	signatureValue, err := runtimeSignature(signer, release.document, issuedAt)
	if err != nil {
		return err
	}
	signature, err := newArtifact("release.sig", map[string]any{
		"contract_version": "release-signature.v1", "manifest_sha256": release.digest,
		"signature": signatureValue,
	})
	if err != nil {
		return err
	}
	for _, name := range []string{"release.json", "release.sig", "trust-root.json"} {
		path := filepath.Join(directory, name)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("replace %s: %w", name, err)
		}
	}
	for _, item := range []artifact{release, signature, signer.trustRoot} {
		if err := publishArtifact(directory, item); err != nil {
			return fmt.Errorf("publish %s: %w", item.name, err)
		}
	}
	return nil
}

func loadReleaseManifest(directory string) (attestation.ReleaseManifest, error) {
	document, err := os.ReadFile(filepath.Join(directory, "release.json"))
	if err != nil {
		return attestation.ReleaseManifest{}, fmt.Errorf("read signed release: %w", err)
	}
	var manifest attestation.ReleaseManifest
	if err := json.Unmarshal(document, &manifest); err != nil {
		return attestation.ReleaseManifest{}, fmt.Errorf("decode signed release: %w", err)
	}
	if strings.TrimSpace(manifest.ReleaseID) == "" ||
		strings.TrimSpace(manifest.DatabaseName) == "" {
		return attestation.ReleaseManifest{}, fmt.Errorf("signed release identity is incomplete")
	}
	return manifest, nil
}
