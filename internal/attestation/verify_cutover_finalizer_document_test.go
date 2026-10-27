package attestation

import (
	"bytes"
	"strings"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestVerifiedCutoverDocumentsCannotBeReplacedByTamperedRawBytes(t *testing.T) {
	root, signer, adoption, adoptionSHA := verifiedCutoverLeaseFixture(t)
	seal, err := VerifyCutoverSessionSeal(
		signCutoverSeal(t, signer, attestationSessionSeal(adoptionSHA)),
		adoption, root, cutoverTestNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	cleanup, err := VerifyCutoverCredentialCleanup(
		signCutoverCleanup(t, signer, attestationCleanup(adoptionSHA, seal.ReceiptSHA256())),
		adoption, seal, root, cutoverTestNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	finalizer := verifiedFinalizerFixture(t, signer, root, adoption, cleanup)
	documents := []struct {
		canonical []byte
		digest    string
		fresh     func() []byte
	}{
		{seal.CanonicalDocument(), seal.ReceiptSHA256(), seal.CanonicalDocument},
		{cleanup.CanonicalDocument(), cleanup.ReceiptSHA256(), cleanup.CanonicalDocument},
		{finalizer.CanonicalDocument(), finalizer.ManifestSHA256(), finalizer.CanonicalDocument},
	}
	for _, document := range documents {
		tampered := document.canonical
		tampered[0] ^= 1
		fresh := document.fresh()
		if bytes.Equal(tampered, fresh) {
			t.Fatal("tampered raw bytes replaced verified canonical document")
		}
		digest, err := DigestHex(fresh)
		if err != nil || digest != document.digest {
			t.Fatalf("verified canonical document binding = %s, %v", digest, err)
		}
	}
}

func verifiedFinalizerFixture(
	t *testing.T,
	signer Signer,
	root TrustRoot,
	adoption VerifiedCutoverCredentialLease,
	cleanup VerifiedCutoverCredentialCleanup,
) VerifiedCutoverFinalizerLease {
	return verifiedFinalizerFixtureForTerminal(
		t, signer, root, adoption, cleanup, strings.Repeat("b", 64), "completed",
	)
}

func verifiedFinalizerFixtureForTerminal(
	t *testing.T,
	signer Signer,
	root TrustRoot,
	adoption VerifiedCutoverCredentialLease,
	cleanup VerifiedCutoverCredentialCleanup,
	terminalDigest string,
	terminalState string,
) VerifiedCutoverFinalizerLease {
	t.Helper()
	var tenantDigest *string
	var tenantCount *int
	if terminalState == "completed" {
		digest := strings.Repeat("d", 64)
		count := 3
		tenantDigest, tenantCount = &digest, &count
	}
	lease := cutoverv1.FinalizerLease{
		ContractVersion: "cutover-finalizer-lease.v1",
		LeaseID:         "99999999-9999-4999-8999-999999999999",
		ReleaseID:       adoption.ReleaseID(), EnvironmentID: adoption.EnvironmentID(),
		OperationID: adoption.OperationID(), RunID: adoption.RunID(),
		IssuedAt: "2026-07-13T12:03:00.001Z", NotBefore: "2026-07-13T12:03:00.001Z",
		ExpiresAt: "2026-07-13T12:05:00.001Z", MaxDurationSeconds: 120,
		RoleLease: cutoverv1.RoleLease{
			Principal: "harpy_cutover_finalizer", Login: true,
			ConnectionLimit: 1, PasswordState: "leased_secret",
		},
		Target: adoption.Target(), AdoptionLeaseSHA256: adoption.ManifestSHA256(),
		CleanupReceiptSHA256:           cleanup.ReceiptSHA256(),
		AuthorityAdoptionReceiptSHA256: strings.Repeat("c", 64),
		FinalizationStateSHA256:        strings.Repeat("a", 64),
		TenantRequirementCatalogSHA256: tenantDigest, TenantRequirementCount: tenantCount,
		FinalizationProcedureOID:       19001,
		FinalizationProcedureSignature: "cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)",
		TerminalState:                  terminalState, TerminalEvidenceSHA256: terminalDigest,
	}
	document := mustReleaseJSON(t, lease)
	signature, err := signer.SignAt(document, cutoverTestNow)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	signed := cutoverv1.FinalizerLeaseSignature{
		ContractVersion: "cutover-finalizer-lease-signature.v1", ManifestSHA256: digest,
		Signature: cutoverv1.Signature{
			Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
		},
	}
	verified, err := VerifyCutoverFinalizerLease(
		document, mustReleaseJSON(t, signed), adoption, cleanup,
		verifiedAuthorityReceiptForFinalizer(adoption), root, cutoverTestNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	return verified
}

func verifiedAuthorityReceiptForFinalizer(
	adoption VerifiedCutoverCredentialLease,
) VerifiedAuthorityAdoptionReceipt {
	return VerifiedAuthorityAdoptionReceipt{
		receipt: cutoverv1.AuthorityAdoptionReceipt{
			ReleaseID: adoption.ReleaseID(), EnvironmentID: adoption.EnvironmentID(),
			Target: adoption.Target(), Result: "succeeded",
			Objects: []cutoverv1.AuthorityAdoptionReceiptObject{{
				SchemaName: "cutover_authority", ObjectName: "finalize_operation_v1",
				IdentityArguments: "uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea",
				ObjectOID:         19001,
			}},
		},
		digest: strings.Repeat("c", 64), verified: true,
	}
}
