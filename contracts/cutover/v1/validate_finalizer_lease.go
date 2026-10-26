package cutoverv1

import "time"

const maximumFinalizerLeaseDuration = 2 * time.Minute

func ValidateFinalizerLease(lease FinalizerLease) error {
	if lease.ContractVersion != "cutover-finalizer-lease.v1" {
		return contractError("finalizer lease contract")
	}
	identifiers := []string{
		lease.LeaseID, lease.ReleaseID, lease.EnvironmentID, lease.OperationID, lease.RunID,
	}
	for _, identifier := range identifiers {
		if !validUUID(identifier) {
			return contractError("finalizer lease identity")
		}
	}
	if !allUnique(identifiers...) || ValidateDatabaseIdentity(lease.Target) != nil ||
		!validDigest(lease.AdoptionLeaseSHA256) || !validDigest(lease.CleanupReceiptSHA256) ||
		!validDigest(lease.AuthorityAdoptionReceiptSHA256) ||
		!validDigest(lease.FinalizationStateSHA256) ||
		!validDigest(lease.TerminalEvidenceSHA256) {
		return contractError("finalizer lease binding")
	}
	issuedAt, issuedErr := parseUTC(lease.IssuedAt)
	notBefore, notBeforeErr := parseUTC(lease.NotBefore)
	expiresAt, expiresErr := parseUTC(lease.ExpiresAt)
	if issuedErr != nil || notBeforeErr != nil || expiresErr != nil ||
		issuedAt.After(notBefore) || !notBefore.Before(expiresAt) ||
		expiresAt.Sub(issuedAt) > maximumFinalizerLeaseDuration ||
		expiresAt.Sub(notBefore) > maximumFinalizerLeaseDuration ||
		lease.MaxDurationSeconds != 120 {
		return contractError("finalizer lease validity window")
	}
	role := lease.RoleLease
	if role.Principal != "harpy_cutover_finalizer" || !role.Login || role.Replication ||
		role.ConnectionLimit != 1 || role.PasswordState != "leased_secret" {
		return contractError("finalizer role lease")
	}
	if lease.FinalizationProcedureOID == 0 ||
		lease.FinalizationProcedureSignature != finalizationProcedureSignature {
		return contractError("finalization procedure")
	}
	if lease.TerminalState != "completed" && lease.TerminalState != "failed" &&
		lease.TerminalState != "cancelled" {
		return contractError("terminal state")
	}
	if lease.TerminalState == "completed" {
		if lease.TenantRequirementCatalogSHA256 == nil ||
			!validDigest(*lease.TenantRequirementCatalogSHA256) ||
			lease.TenantRequirementCount == nil ||
			*lease.TenantRequirementCount < 1 || *lease.TenantRequirementCount > 10000 {
			return contractError("tenant requirement authority")
		}
	} else if lease.TenantRequirementCatalogSHA256 != nil || lease.TenantRequirementCount != nil {
		return contractError("non-completed tenant requirement authority")
	}
	return nil
}

func ValidateFinalizerLeaseSignature(value FinalizerLeaseSignature) error {
	if value.ContractVersion != "cutover-finalizer-lease-signature.v1" ||
		!validDigest(value.ManifestSHA256) {
		return contractError("finalizer lease signature contract")
	}
	return validateSignature(value.Signature)
}

const finalizationProcedureSignature = "cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)"
