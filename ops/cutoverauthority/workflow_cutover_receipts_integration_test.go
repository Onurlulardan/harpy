//go:build integration

package main

import (
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func workflowSessionSeal(
	t *testing.T, signer workflowSigner, lease cutoverv1.CredentialLease,
	leaseSHA string, ids workflowCutoverIDs, now time.Time,
) []byte {
	t.Helper()
	application := "harpy-cutover/" + ids.run
	receipt := cutoverv1.SessionSealReceipt{
		ContractVersion: "cutover-session-seal-receipt.v1", Kind: "cutover_session_seal",
		Payload: cutoverv1.SessionSealPayload{
			ReceiptID: ids.seal, LeaseID: ids.lease, LeaseManifestSHA256: leaseSHA,
			EnvironmentID: workflowEnvironmentID, OperationID: ids.operation,
			RunID: ids.run, IssuedAt: now.Add(-time.Minute).Format(time.RFC3339),
			Source: lease.Source, Target: lease.Target,
			Credentials: []cutoverv1.SealedCredential{
				{Principal: "harpy_cutover_decoder", ConnectionLimit: 0},
				{Principal: "harpy_adopter", ConnectionLimit: 0},
			},
			SourceSessions: []cutoverv1.CutoverSession{
				workflowSession(101, now.Add(-90*time.Second), "harpy_cutover_decoder", application),
				workflowSession(102, now.Add(-80*time.Second), "harpy_cutover_decoder", application),
			},
			TargetSessions: []cutoverv1.CutoverSession{
				workflowSession(201, now.Add(-70*time.Second), "harpy_adopter", application),
			}, Result: "sealed",
		},
	}
	return signWorkflowEnvelope(t, signer, receipt.SigningEnvelope(),
		now.Add(-time.Minute), func(signature cutoverv1.Signature) any {
			receipt.Signature = signature
			return receipt
		})
}

func workflowSession(
	pid int64, started time.Time, principal, application string,
) cutoverv1.CutoverSession {
	return cutoverv1.CutoverSession{PID: pid, BackendStart: started.Format(time.RFC3339),
		SessionUser: principal, ApplicationName: application}
}

func workflowCleanup(
	t *testing.T, signer workflowSigner, lease cutoverv1.CredentialLease,
	leaseSHA, sealSHA string, ids workflowCutoverIDs, now time.Time,
) []byte {
	t.Helper()
	receipt := cutoverv1.CredentialCleanupReceipt{
		ContractVersion: "cutover-credential-cleanup-receipt.v1",
		Kind:            "cutover_credential_cleanup", Payload: cutoverv1.CredentialCleanupPayload{
			ReceiptID: ids.cleanup, LeaseID: ids.lease, LeaseManifestSHA256: leaseSHA,
			SealReceiptSHA256: sealSHA, EnvironmentID: workflowEnvironmentID,
			OperationID: ids.operation, RunID: ids.run,
			IssuedAt: now.Add(-30 * time.Second).Format(time.RFC3339),
			Principals: []cutoverv1.CleanedPrincipal{
				workflowCleanedPrincipal("harpy_cutover_decoder", now.Add(-31*time.Second)),
				workflowCleanedPrincipal("harpy_adopter", now.Add(-30*time.Second)),
			}, Result: "succeeded",
		},
	}
	return signWorkflowEnvelope(t, signer, receipt.SigningEnvelope(),
		now.Add(-30*time.Second), func(signature cutoverv1.Signature) any {
			receipt.Signature = signature
			return receipt
		})
}

func workflowCleanedPrincipal(
	principal string, cleanedAt time.Time,
) cutoverv1.CleanedPrincipal {
	return cutoverv1.CleanedPrincipal{Principal: principal,
		CleanedAt: cleanedAt.Format(time.RFC3339), ConnectionLimit: 0,
		DynamicGrantsRevoked: true}
}

func signWorkflowEnvelope(
	t *testing.T, signer workflowSigner, envelope any, issuedAt time.Time,
	complete func(cutoverv1.Signature) any,
) []byte {
	t.Helper()
	document, err := attestation.MarshalCanonical(envelope)
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.signer.SignAt(document, issuedAt)
	if err != nil {
		t.Fatal(err)
	}
	completed, err := attestation.MarshalCanonical(complete(cutoverv1.Signature{
		Algorithm: signed.Algorithm, KeyID: signed.KeyID, Value: signed.Value,
	}))
	if err != nil {
		t.Fatal(err)
	}
	return completed
}
