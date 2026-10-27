package cutoverv1

func ValidateCredentialCleanupReceipt(receipt CredentialCleanupReceipt) error {
	if receipt.ContractVersion != "cutover-credential-cleanup-receipt.v1" ||
		receipt.Kind != "cutover_credential_cleanup" || receipt.Payload.Result != "succeeded" {
		return contractError("credential cleanup contract")
	}
	payload := receipt.Payload
	identifiers := []string{
		payload.ReceiptID, payload.LeaseID, payload.EnvironmentID, payload.OperationID, payload.RunID,
	}
	for _, identifier := range identifiers {
		if !validUUID(identifier) {
			return contractError("credential cleanup identity")
		}
	}
	issuedAt, issuedErr := parseUTC(payload.IssuedAt)
	if !allUnique(identifiers...) || !validDigest(payload.LeaseManifestSHA256) ||
		!validDigest(payload.SealReceiptSHA256) || issuedErr != nil || len(payload.Principals) != 2 {
		return contractError("credential cleanup binding")
	}
	expected := []string{"harpy_cutover_decoder", "harpy_adopter"}
	for index, principal := range payload.Principals {
		cleanedAt, err := parseUTC(principal.CleanedAt)
		if err != nil || cleanedAt.After(issuedAt) || principal.Principal != expected[index] ||
			principal.CanLogin || principal.PasswordState != nil || principal.Replication ||
			principal.ConnectionLimit != 0 || principal.ActiveSessions != 0 ||
			principal.MembershipCount != 0 || !principal.DynamicGrantsRevoked {
			return contractError("cleaned principal state")
		}
	}
	return validateSignature(receipt.Signature)
}

func ValidateCredentialCleanupBinding(
	receipt CredentialCleanupReceipt,
	lease CredentialLease,
	leaseManifestSHA256 string,
	seal SessionSealReceipt,
	sealReceiptSHA256 string,
) error {
	if ValidateCredentialCleanupReceipt(receipt) != nil || ValidateCredentialLease(lease) != nil ||
		ValidateSessionSealBinding(seal, lease, leaseManifestSHA256) != nil ||
		!validDigest(sealReceiptSHA256) {
		return contractError("credential cleanup verification input")
	}
	payload := receipt.Payload
	if payload.LeaseID != lease.LeaseID || payload.LeaseManifestSHA256 != leaseManifestSHA256 ||
		payload.SealReceiptSHA256 != sealReceiptSHA256 ||
		payload.EnvironmentID != lease.EnvironmentID || payload.OperationID != lease.OperationID ||
		payload.RunID != lease.RunID || payload.ReceiptID == seal.Payload.ReceiptID ||
		payload.ReceiptID == lease.ReleaseID || payload.ReceiptID == lease.ScopeID {
		return contractError("credential cleanup lease binding")
	}
	issuedAt, _ := parseUTC(payload.IssuedAt)
	sealIssuedAt, _ := parseUTC(seal.Payload.IssuedAt)
	if issuedAt.Before(sealIssuedAt) {
		return contractError("credential cleanup issue sequence")
	}
	for _, principal := range payload.Principals {
		cleanedAt, _ := parseUTC(principal.CleanedAt)
		if cleanedAt.Before(sealIssuedAt) {
			return contractError("credential cleanup sequence")
		}
	}
	return nil
}
