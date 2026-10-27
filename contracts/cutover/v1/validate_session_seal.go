package cutoverv1

import (
	"fmt"
	"time"
)

func ValidateSessionSealReceipt(receipt SessionSealReceipt) error {
	if receipt.ContractVersion != "cutover-session-seal-receipt.v1" ||
		receipt.Kind != "cutover_session_seal" || receipt.Payload.Result != "sealed" {
		return contractError("session seal contract")
	}
	payload := receipt.Payload
	identifiers := []string{
		payload.ReceiptID, payload.LeaseID, payload.EnvironmentID, payload.OperationID, payload.RunID,
	}
	for _, identifier := range identifiers {
		if !validUUID(identifier) {
			return contractError("session seal identity")
		}
	}
	if !allUnique(identifiers...) || !validDigest(payload.LeaseManifestSHA256) ||
		ValidateDatabaseIdentity(payload.Source) != nil ||
		ValidateDatabaseIdentity(payload.Target) != nil ||
		!sameDatabaseIdentity(payload.Source, payload.Target) {
		return contractError("session seal binding")
	}
	issuedAt, err := parseUTC(payload.IssuedAt)
	if err != nil || validateSealedCredentials(payload.Credentials) != nil {
		return contractError("session seal credential state")
	}
	applicationName := fmt.Sprintf("harpy-cutover/%s", payload.RunID)
	seen := make(map[int64]struct{}, len(payload.SourceSessions)+len(payload.TargetSessions))
	if err := validateSessionSet(
		payload.SourceSessions, 2, "harpy_cutover_decoder", applicationName, issuedAt, seen,
	); err != nil {
		return err
	}
	if err := validateSessionSet(
		payload.TargetSessions, 1, "harpy_adopter", applicationName, issuedAt, seen,
	); err != nil {
		return err
	}
	return validateSignature(receipt.Signature)
}

func validateSealedCredentials(credentials []SealedCredential) error {
	if len(credentials) != 2 {
		return contractError("sealed credential set")
	}
	expected := []string{"harpy_cutover_decoder", "harpy_adopter"}
	for index, credential := range credentials {
		if credential.Principal != expected[index] || credential.CanLogin ||
			credential.PasswordState != nil || credential.Replication || credential.ConnectionLimit != 0 {
			return contractError("sealed credential")
		}
	}
	return nil
}

func validateSessionSet(
	sessions []CutoverSession,
	minimum int,
	principal string,
	applicationName string,
	issuedAt time.Time,
	seen map[int64]struct{},
) error {
	if len(sessions) < minimum || len(sessions) > 32 {
		return contractError("sealed session count")
	}
	for _, session := range sessions {
		backendStart, err := parseUTC(session.BackendStart)
		if err != nil || session.PID < 1 || session.SessionUser != principal ||
			session.ApplicationName != applicationName || backendStart.After(issuedAt) {
			return contractError("sealed session")
		}
		if _, duplicate := seen[session.PID]; duplicate {
			return contractError("duplicate sealed session")
		}
		seen[session.PID] = struct{}{}
	}
	return nil
}

func ValidateSessionSealBinding(
	receipt SessionSealReceipt,
	lease CredentialLease,
	leaseManifestSHA256 string,
) error {
	if ValidateSessionSealReceipt(receipt) != nil || ValidateCredentialLease(lease) != nil ||
		!validDigest(leaseManifestSHA256) {
		return contractError("session seal verification input")
	}
	payload := receipt.Payload
	if payload.LeaseID != lease.LeaseID || payload.LeaseManifestSHA256 != leaseManifestSHA256 ||
		payload.EnvironmentID != lease.EnvironmentID || payload.OperationID != lease.OperationID ||
		payload.RunID != lease.RunID || !sameDatabaseIdentity(payload.Source, lease.Source) ||
		!sameDatabaseIdentity(payload.Target, lease.Target) ||
		payload.ReceiptID == lease.ReleaseID || payload.ReceiptID == lease.ScopeID ||
		len(payload.SourceSessions) > lease.SourceConnectionLimit ||
		len(payload.TargetSessions) > lease.TargetConnectionLimit {
		return contractError("session seal lease binding")
	}
	issuedAt, _ := parseUTC(payload.IssuedAt)
	notBefore, _ := parseUTC(lease.NotBefore)
	expiresAt, _ := parseUTC(lease.ExpiresAt)
	if issuedAt.Before(notBefore) || !issuedAt.Before(expiresAt) {
		return contractError("session seal lease window")
	}
	for _, sessions := range [][]CutoverSession{payload.SourceSessions, payload.TargetSessions} {
		for _, session := range sessions {
			backendStart, _ := parseUTC(session.BackendStart)
			if backendStart.Before(notBefore) {
				return contractError("session start lease window")
			}
		}
	}
	return nil
}
