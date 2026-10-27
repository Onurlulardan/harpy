package runtimev1

import "fmt"

func ValidateBootstrapAuthorityManifest(manifest BootstrapAuthorityManifest) error {
	if manifest.ContractVersion != "bootstrap-authority-manifest.v1" {
		return fmt.Errorf("bootstrap authority contract version is invalid")
	}
	if err := validateUUID(manifest.ManifestID, "manifest_id"); err != nil {
		return err
	}
	if err := validateUUID(manifest.EnvironmentID, "environment_id"); err != nil {
		return err
	}
	if err := validateUUID(manifest.ReleaseID, "release_id"); err != nil {
		return err
	}
	issuedAt, err := validateTimestamp(manifest.IssuedAt, "issued_at")
	if err != nil {
		return err
	}
	expiresAt, err := validateTimestamp(manifest.ExpiresAt, "expires_at")
	if err != nil || !issuedAt.Before(expiresAt) {
		return fmt.Errorf("bootstrap authority validity window is invalid")
	}
	origin := DatabaseOrigin(manifest.DatabaseOrigin)
	if !origin.Valid() {
		return fmt.Errorf("database_origin is invalid")
	}
	if err := validateBootstrapAuthorityDatabase(manifest.Database, origin); err != nil {
		return err
	}
	if manifest.Connections.EndpointKind != "direct_postgres" ||
		manifest.Connections.MaintenanceDatabase != "postgres" ||
		manifest.Connections.TargetDatabase != manifest.Database.Name ||
		!manifest.Connections.SameServerRequired {
		return fmt.Errorf("bootstrap authority connection contract is invalid")
	}
	if manifest.CreateRoleSelfGrant != "" {
		return fmt.Errorf("bootstrap authority createrole self-grant is invalid")
	}
	if err := validateBootstrapAuthorityRoles(manifest.Roles, origin, manifest.Database); err != nil {
		return err
	}
	if err := validateBootstrapAuthorityMemberships(manifest.Memberships, origin); err != nil {
		return err
	}
	if err := validateBootstrapCredentialLeases(
		manifest.CredentialLeases, issuedAt, expiresAt,
	); err != nil {
		return err
	}
	return validateBootstrapAuthorityDigests(manifest)
}

func validateBootstrapAuthorityDigests(manifest BootstrapAuthorityManifest) error {
	digests := []struct {
		value string
		field string
	}{
		{manifest.RolesManifestSHA256, "roles_manifest_sha256"},
		{manifest.PrivilegesManifestSHA256, "privileges_manifest_sha256"},
		{manifest.OwnershipManifestSHA256, "ownership_manifest_sha256"},
		{manifest.BootstrapOperationDigest, "bootstrap_operation_digest"},
		{manifest.CredentialCleanupOperationDigest, "credential_cleanup_operation_digest"},
	}
	for _, digest := range digests {
		if err := validateDigest(digest.value, digest.field); err != nil {
			return err
		}
	}
	return nil
}

func ValidateBootstrapAuthoritySignature(signature BootstrapAuthoritySignature) error {
	if signature.ContractVersion != "bootstrap-authority-signature.v1" {
		return fmt.Errorf("bootstrap authority signature contract version is invalid")
	}
	if err := validateDigest(signature.ManifestSHA256, "manifest_sha256"); err != nil {
		return err
	}
	return ValidateAttestationSignature(signature.Signature)
}
