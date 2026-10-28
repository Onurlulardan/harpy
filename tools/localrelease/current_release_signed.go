package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const currentSignedReleasePath = "/run/release/current/release.json"

func verifyCurrentSignedRelease(now time.Time) error {
	document, err := os.ReadFile(currentSignedReleasePath)
	if err != nil {
		return fmt.Errorf("current signed release is absent")
	}
	var manifest attestation.ReleaseManifest
	if err := json.Unmarshal(document, &manifest); err != nil {
		return fmt.Errorf("current signed release is invalid")
	}
	expiresAt, err := time.Parse(time.RFC3339, manifest.ExpiresAt)
	if err != nil || !now.UTC().Before(expiresAt) {
		return fmt.Errorf("current signed release has expired")
	}
	return nil
}
