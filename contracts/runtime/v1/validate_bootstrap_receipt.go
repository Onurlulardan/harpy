package runtimev1

import "fmt"

func ValidateBootstrapReceipt(receipt BootstrapReceipt) error {
	payload := receipt.Payload
	if receipt.ContractVersion != "postgres-bootstrap-receipt.v1" ||
		receipt.Kind != "cluster_bootstrap" {
		return fmt.Errorf("bootstrap receipt identity is invalid")
	}
	for _, identity := range []struct {
		value string
		field string
	}{
		{payload.RunID, "run_id"}, {payload.ReleaseID, "release_id"},
		{payload.EnvironmentID, "environment_id"},
		{payload.CredentialLeaseID, "credential_lease_id"},
	} {
		if err := validateUUID(identity.value, identity.field); err != nil {
			return err
		}
	}
	if _, err := validateTimestamp(payload.IssuedAt, "issued_at"); err != nil {
		return err
	}
	if err := validateReceiptDigests(
		receiptDigest{payload.ReleaseManifestSHA256, "release_manifest_sha256"},
		receiptDigest{payload.BootstrapAuthorityManifestSHA256, "bootstrap_authority_manifest_sha256"},
		receiptDigest{payload.OwnershipManifestSHA256, "ownership_manifest_sha256"},
		receiptDigest{payload.OperationDigest, "operation_digest"},
		receiptDigest{payload.RolesManifestSHA256, "roles_manifest_sha256"},
		receiptDigest{payload.PrivilegesManifestSHA256, "privileges_manifest_sha256"},
	); err != nil {
		return err
	}
	if _, err := validateReceiptDatabase(
		payload.DatabaseName, payload.DatabaseSystemIdentifier, payload.DatabaseOrigin,
		payload.DatabaseOwner, payload.DatabaseEncoding, payload.DatabaseLocaleProvider,
		payload.DatabaseCollation, payload.DatabaseCType, payload.DatabaseOID,
		payload.DataChecksums,
	); err != nil {
		return err
	}
	if payload.PostgresImageDigest != PostgresImageDigest ||
		payload.ServerVersionNumber != PostgresServerVersionNumber ||
		payload.SessionUser != "harpy_cluster_admin" ||
		payload.CurrentUser != "harpy_cluster_admin" || payload.Result != "succeeded" {
		return fmt.Errorf("bootstrap receipt execution contract is invalid")
	}
	return ValidateAttestationSignature(receipt.Signature)
}
