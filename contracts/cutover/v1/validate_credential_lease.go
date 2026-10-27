package cutoverv1

import "time"

const maximumLeaseDuration = 15 * time.Minute

func ValidateCredentialLease(lease CredentialLease) error {
	if lease.ContractVersion != "cutover-credential-lease.v1" {
		return contractError("credential lease contract")
	}
	if !validLeaseIdentifiers(lease) || !validDigest(lease.PlanSHA256) ||
		!validDigest(lease.TransformManifestSHA256) {
		return contractError("credential lease identity")
	}
	if err := validateLeaseWindow(lease); err != nil {
		return err
	}
	if err := validateLeaseScope(lease); err != nil {
		return err
	}
	if ValidateDatabaseIdentity(lease.Source) != nil ||
		ValidateDatabaseIdentity(lease.Target) != nil ||
		!sameDatabaseIdentity(lease.Source, lease.Target) {
		return contractError("credential lease database binding")
	}
	if !sqlNamePattern.MatchString(lease.PublicationName) ||
		!sqlNamePattern.MatchString(lease.SlotName) || lease.PublicationName == lease.SlotName {
		return contractError("publication or slot identity")
	}
	if lease.SourceConnectionLimit < 2 || lease.SourceConnectionLimit > 32 ||
		lease.TargetConnectionLimit < 1 || lease.TargetConnectionLimit > 32 {
		return contractError("connection bounds")
	}
	if err := validateRoleLeases(lease); err != nil {
		return err
	}
	return validateRelationMappings(lease.Relations, lease.Source, lease.Target)
}

func validLeaseIdentifiers(lease CredentialLease) bool {
	identifiers := []string{
		lease.LeaseID, lease.ReleaseID, lease.EnvironmentID, lease.OperationID, lease.RunID,
	}
	for _, identifier := range identifiers {
		if !validUUID(identifier) {
			return false
		}
	}
	if !allUnique(identifiers...) || !validUUID(lease.ScopeID) {
		return false
	}
	if lease.ScopeKind == "global" {
		return lease.ScopeID == lease.EnvironmentID
	}
	return lease.ScopeKind == "tenant" && allUnique(append(identifiers, lease.ScopeID)...)
}

func validateLeaseWindow(lease CredentialLease) error {
	issuedAt, issuedErr := parseUTC(lease.IssuedAt)
	notBefore, notBeforeErr := parseUTC(lease.NotBefore)
	expiresAt, expiresErr := parseUTC(lease.ExpiresAt)
	if issuedErr != nil || notBeforeErr != nil || expiresErr != nil ||
		issuedAt.After(notBefore) || !notBefore.Before(expiresAt) ||
		expiresAt.Sub(issuedAt) > maximumLeaseDuration ||
		expiresAt.Sub(notBefore) > maximumLeaseDuration || lease.MaxDurationSeconds != 900 {
		return contractError("credential lease validity window")
	}
	return nil
}

func validateRoleLeases(lease CredentialLease) error {
	if len(lease.RoleLeases) != 2 {
		return contractError("role lease set")
	}
	decoder, adopter := lease.RoleLeases[0], lease.RoleLeases[1]
	if decoder.Principal != "harpy_cutover_decoder" || !decoder.Login ||
		!decoder.Replication || decoder.ConnectionLimit != lease.SourceConnectionLimit ||
		decoder.PasswordState != "leased_secret" {
		return contractError("decoder role lease")
	}
	if adopter.Principal != "harpy_adopter" || !adopter.Login || adopter.Replication ||
		adopter.ConnectionLimit != lease.TargetConnectionLimit ||
		adopter.PasswordState != "leased_secret" {
		return contractError("adopter role lease")
	}
	return nil
}
