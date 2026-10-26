package cutoverv1

import (
	"strings"
	"testing"
)

func TestValidateSessionSealBindingAcceptsExactSessions(t *testing.T) {
	leaseSHA := strings.Repeat("a", 64)
	if err := ValidateSessionSealBinding(
		validTypedSessionSeal(leaseSHA), validTypedCredentialLease(), leaseSHA,
	); err != nil {
		t.Fatalf("valid session seal rejected: %v", err)
	}
}

func TestValidateSessionSealRejectsSessionOrLeaseDrift(t *testing.T) {
	leaseSHA := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		mutate func(*SessionSealReceipt)
	}{
		{"wrong application run", func(v *SessionSealReceipt) {
			v.Payload.SourceSessions[0].ApplicationName = "harpy-cutover/99999999-9999-4999-8999-999999999999"
		}},
		{"missing decoder session", func(v *SessionSealReceipt) { v.Payload.SourceSessions = v.Payload.SourceSessions[:1] }},
		{"duplicate pid", func(v *SessionSealReceipt) { v.Payload.TargetSessions[0].PID = v.Payload.SourceSessions[0].PID }},
		{"session starts after seal", func(v *SessionSealReceipt) { v.Payload.TargetSessions[0].BackendStart = "2026-07-13T12:02:01Z" }},
		{"credential password not null", func(v *SessionSealReceipt) {
			state := "retained"
			v.Payload.Credentials[0].PasswordState = &state
		}},
		{"database mismatch", func(v *SessionSealReceipt) { v.Payload.Target.DatabaseOID++ }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			receipt := validTypedSessionSeal(leaseSHA)
			test.mutate(&receipt)
			if err := ValidateSessionSealBinding(receipt, validTypedCredentialLease(), leaseSHA); err == nil {
				t.Fatal("unsafe session seal accepted")
			}
		})
	}
}

func TestValidateCredentialCleanupBindingAcceptsTerminalEvidence(t *testing.T) {
	leaseSHA := strings.Repeat("a", 64)
	sealSHA := strings.Repeat("b", 64)
	err := ValidateCredentialCleanupBinding(
		validTypedCleanup(leaseSHA, sealSHA), validTypedCredentialLease(), leaseSHA,
		validTypedSessionSeal(leaseSHA), sealSHA,
	)
	if err != nil {
		t.Fatalf("valid cleanup receipt rejected: %v", err)
	}
}

func TestValidateCredentialCleanupRejectsNonTerminalState(t *testing.T) {
	leaseSHA := strings.Repeat("a", 64)
	sealSHA := strings.Repeat("b", 64)
	tests := []struct {
		name   string
		mutate func(*CredentialCleanupReceipt)
	}{
		{"active session remains", func(v *CredentialCleanupReceipt) { v.Payload.Principals[0].ActiveSessions = 1 }},
		{"membership remains", func(v *CredentialCleanupReceipt) { v.Payload.Principals[1].MembershipCount = 1 }},
		{"grant remains", func(v *CredentialCleanupReceipt) { v.Payload.Principals[0].DynamicGrantsRevoked = false }},
		{"password not null", func(v *CredentialCleanupReceipt) {
			state := "retained"
			v.Payload.Principals[1].PasswordState = &state
		}},
		{"cleanup precedes seal", func(v *CredentialCleanupReceipt) { v.Payload.Principals[0].CleanedAt = "2026-07-13T12:01:59Z" }},
		{"wrong seal digest", func(v *CredentialCleanupReceipt) { v.Payload.SealReceiptSHA256 = strings.Repeat("c", 64) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			receipt := validTypedCleanup(leaseSHA, sealSHA)
			test.mutate(&receipt)
			err := ValidateCredentialCleanupBinding(
				receipt, validTypedCredentialLease(), leaseSHA,
				validTypedSessionSeal(leaseSHA), sealSHA,
			)
			if err == nil {
				t.Fatal("non-terminal cleanup receipt accepted")
			}
		})
	}
}
