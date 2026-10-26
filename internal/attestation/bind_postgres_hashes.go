package attestation

import (
	"crypto/subtle"
	"fmt"
)

func bindReleaseArtifactHashes(release ReleaseManifest, artifacts PostgresArtifactSet) error {
	bindings := []struct {
		expected string
		actual   string
		name     string
	}{
		{release.MigrationManifestSHA256, artifacts.Migration.DigestSHA256(), "migration"},
		{release.ExtensionManifestSHA256, artifacts.Extension.DigestSHA256(), "extension"},
		{release.RolesManifestSHA256, artifacts.Roles.DigestSHA256(), "roles"},
		{release.PrivilegesManifestSHA256, artifacts.Privileges.DigestSHA256(), "privileges"},
		{release.OwnershipManifestSHA256, artifacts.Ownership.Root().DigestSHA256(), "ownership"},
		{release.BootstrapOperationDigest, artifacts.BootstrapOperation.DigestSHA256(), "bootstrap operation"},
		{release.ExtensionOperationDigest, artifacts.ExtensionOperation.DigestSHA256(), "extension operation"},
		{release.CredentialCleanupDigest, artifacts.CleanupOperation.DigestSHA256(), "cleanup operation"},
	}
	for _, binding := range bindings {
		if !sameDigest(binding.expected, binding.actual) {
			return fmt.Errorf("%s digest differs from release", binding.name)
		}
	}
	return nil
}

func bindAuthorityArtifacts(
	release VerifiedRelease,
	authority VerifiedBootstrapAuthority,
	artifacts PostgresArtifactSet,
) error {
	manifest := authority.Value()
	releaseManifest := release.Value()
	if manifest.ReleaseID != releaseManifest.ReleaseID ||
		manifest.DatabaseOrigin != releaseManifest.DatabaseOrigin ||
		manifest.Database.Name != releaseManifest.DatabaseName ||
		!sameDigest(releaseManifest.BootstrapAuthoritySHA256, authority.ManifestSHA256()) {
		return fmt.Errorf("bootstrap authority identity differs from release")
	}
	bindings := [][2]string{
		{manifest.RolesManifestSHA256, artifacts.Roles.DigestSHA256()},
		{manifest.PrivilegesManifestSHA256, artifacts.Privileges.DigestSHA256()},
		{manifest.OwnershipManifestSHA256, artifacts.Ownership.Root().DigestSHA256()},
		{manifest.BootstrapOperationDigest, artifacts.BootstrapOperation.DigestSHA256()},
		{manifest.CredentialCleanupOperationDigest, artifacts.CleanupOperation.DigestSHA256()},
	}
	for _, binding := range bindings {
		if !sameDigest(binding[0], binding[1]) {
			return fmt.Errorf("artifact digest differs from bootstrap authority")
		}
	}
	return nil
}

func sameDigest(left, right string) bool {
	return len(left) == 64 && len(right) == 64 &&
		subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
