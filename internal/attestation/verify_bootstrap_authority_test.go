package attestation

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestVerifyBootstrapAuthorityReturnsOpaqueVerifiedValue(t *testing.T) {
	root, signer := runtimeTestRootAndSigner(t)
	manifest := validRuntimeAuthority()
	manifestDocument := mustRuntimeJSON(t, manifest)
	signatureDocument := signRuntimeAuthority(t, signer, manifestDocument, manifest.IssuedAt)
	verified, err := VerifyBootstrapAuthority(
		manifestDocument, signatureDocument, root, runtimeTestNow,
	)
	if err != nil {
		t.Fatalf("verify bootstrap authority: %v", err)
	}
	if verified.ManifestSHA256() == "" || verified.Value().EnvironmentID != manifest.EnvironmentID {
		t.Fatal("verified authority omitted stable identity or digest")
	}
	mutable := verified.Value()
	mutable.Roles[0].Name = "tampered"
	if verified.Value().Roles[0].Name != "harpy_owner" {
		t.Fatal("verified authority exposed mutable role storage")
	}
}

func TestVerifyBootstrapAuthorityRejectsTamperAndExpiry(t *testing.T) {
	root, signer := runtimeTestRootAndSigner(t)
	manifest := validRuntimeAuthority()
	manifestDocument := mustRuntimeJSON(t, manifest)
	signatureDocument := signRuntimeAuthority(t, signer, manifestDocument, manifest.IssuedAt)
	manifest.EnvironmentID = "99999999-9999-4999-8999-999999999999"
	if _, err := VerifyBootstrapAuthority(
		mustRuntimeJSON(t, manifest), signatureDocument, root, runtimeTestNow,
	); err == nil {
		t.Fatal("tampered authority manifest accepted")
	}
	if _, err := VerifyBootstrapAuthority(
		manifestDocument, signatureDocument, root,
		time.Date(2026, 7, 13, 0, 15, 0, 0, time.UTC),
	); err == nil {
		t.Fatal("expired authority manifest accepted")
	}
}

func signRuntimeAuthority(
	t *testing.T,
	signer Signer,
	manifestDocument []byte,
	issuedAtValue string,
) []byte {
	t.Helper()
	issuedAt, err := time.Parse(time.RFC3339Nano, issuedAtValue)
	if err != nil {
		t.Fatalf("parse authority issue time: %v", err)
	}
	signature, err := signer.SignAt(manifestDocument, issuedAt)
	if err != nil {
		t.Fatalf("sign authority: %v", err)
	}
	digest, err := DigestHex(manifestDocument)
	if err != nil {
		t.Fatalf("digest authority: %v", err)
	}
	return mustRuntimeJSON(t, runtimev1.BootstrapAuthoritySignature{
		ContractVersion: "bootstrap-authority-signature.v1",
		ManifestSHA256:  digest,
		Signature:       runtimeContractSignature(signature),
	})
}

func runtimeContractSignature(signature Signature) runtimev1.AttestationSignature {
	return runtimev1.AttestationSignature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}
}
