package runtimev1

import "fmt"

func ValidateCredentialCleanupReceipt(receipt CredentialCleanupReceipt) error {
	payload := receipt.Payload
	if receipt.ContractVersion != "credential-cleanup-receipt.v1" ||
		receipt.Kind != "credential_cleanup" {
		return fmt.Errorf("credential cleanup receipt identity is invalid")
	}
	if err := validateUUID(payload.RunID, "run_id"); err != nil {
		return err
	}
	if err := validateUUID(payload.ReleaseID, "release_id"); err != nil {
		return err
	}
	issuedAt, err := validateTimestamp(payload.IssuedAt, "issued_at")
	if err != nil {
		return err
	}
	if err := validateReceiptDigests(
		receiptDigest{payload.ReleaseManifestSHA256, "release_manifest_sha256"},
		receiptDigest{payload.BootstrapAuthorityManifestSHA256, "bootstrap_authority_manifest_sha256"},
		receiptDigest{payload.CleanupOperationDigest, "cleanup_operation_digest"},
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
	if payload.Result != "succeeded" || len(payload.Principals) != 2 {
		return fmt.Errorf("credential cleanup result is invalid")
	}
	expected := []string{"harpy_cluster_admin", "harpy_extension_admin"}
	leaseIDs := make(map[string]struct{}, len(payload.Principals))
	for index, principal := range payload.Principals {
		if principal.Name != expected[index] || principal.CanLogin ||
			principal.PasswordState != nil || principal.ActiveSessions != 0 ||
			principal.MembershipCount != 0 {
			return fmt.Errorf("cleaned principal %d retains authority", index)
		}
		if err := validateUUID(principal.LeaseID, "principal.lease_id"); err != nil {
			return err
		}
		if _, duplicate := leaseIDs[principal.LeaseID]; duplicate {
			return fmt.Errorf("cleaned principal lease IDs must be unique")
		}
		leaseIDs[principal.LeaseID] = struct{}{}
		cleanedAt, err := validateTimestamp(principal.CleanedAt, "principal.cleaned_at")
		if err != nil || cleanedAt.After(issuedAt) {
			return fmt.Errorf("cleaned principal %d timestamp is invalid", index)
		}
	}
	return ValidateAttestationSignature(receipt.Signature)
}
