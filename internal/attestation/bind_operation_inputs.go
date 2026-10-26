package attestation

import "fmt"

func bindOperationInputs(artifacts PostgresArtifactSet) error {
	bootstrap := artifacts.BootstrapOperation.Value().InputHashes
	if !sameDigest(bootstrap.RolesManifestSHA256, artifacts.Roles.DigestSHA256()) ||
		!sameDigest(bootstrap.PrivilegesManifestSHA256, artifacts.Privileges.DigestSHA256()) ||
		!sameDigest(bootstrap.OwnershipManifestSHA256, artifacts.Ownership.Root().DigestSHA256()) {
		return fmt.Errorf("bootstrap operation input hashes differ")
	}
	extension := artifacts.ExtensionOperation.Value().InputHashes
	if !sameDigest(extension.ExtensionManifestSHA256, artifacts.Extension.DigestSHA256()) ||
		!sameDigest(extension.OwnershipManifestSHA256, artifacts.Ownership.Root().DigestSHA256()) {
		return fmt.Errorf("extension operation input hashes differ")
	}
	cleanup := artifacts.CleanupOperation.Value().InputHashes
	if !sameDigest(cleanup.RolesManifestSHA256, artifacts.Roles.DigestSHA256()) ||
		!sameDigest(cleanup.PrivilegesManifestSHA256, artifacts.Privileges.DigestSHA256()) {
		return fmt.Errorf("credential cleanup operation input hashes differ")
	}
	return nil
}
