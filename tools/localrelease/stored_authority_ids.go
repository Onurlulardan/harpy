package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/google/uuid"
)

func releaseIDsForDirectory(directory string) releaseIDs {
	manifest, ok := readStoredAuthorityManifest(
		filepath.Join(directory, "bootstrap-authority.json"),
	)
	if !ok {
		return newReleaseIDs()
	}
	ids := newReleaseIDs()
	ids.environment = manifest.EnvironmentID
	ids.release = manifest.ReleaseID
	ids.authorityManifest = manifest.ManifestID
	return ids
}

func readStoredAuthorityManifest(path string) (runtimev1.BootstrapAuthorityManifest, bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return runtimev1.BootstrapAuthorityManifest{}, false
	}
	var manifest runtimev1.BootstrapAuthorityManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return runtimev1.BootstrapAuthorityManifest{}, false
	}
	if err := runtimev1.ValidateBootstrapAuthorityManifest(manifest); err != nil {
		return runtimev1.BootstrapAuthorityManifest{}, false
	}
	if !isUUIDv7(manifest.EnvironmentID) ||
		!isUUIDv7(manifest.ReleaseID) ||
		!isUUIDv7(manifest.ManifestID) {
		return runtimev1.BootstrapAuthorityManifest{}, false
	}
	return manifest, true
}

func isUUIDv7(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.Version() == uuid.Version(7)
}
