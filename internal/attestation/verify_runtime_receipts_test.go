package attestation

import (
	"bytes"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestVerifyRuntimeReceiptsAcceptsSignedExactContracts(t *testing.T) {
	root, signer := runtimeTestRootAndSigner(t)
	bootstrapDocument := signRuntimeBootstrapReceipt(t, signer, validRuntimeBootstrapReceipt())
	bootstrap, err := VerifyBootstrapReceipt(bootstrapDocument, root, runtimeTestNow)
	if err != nil || bootstrap.Value().Payload.EnvironmentID == "" {
		t.Fatalf("verify bootstrap receipt: %v", err)
	}
	extensionDocument := signRuntimeExtensionReceipt(t, signer, validRuntimeExtensionReceipt())
	extension, err := VerifyExtensionReceipt(extensionDocument, root, runtimeTestNow)
	if err != nil {
		t.Fatalf("verify extension receipt: %v", err)
	}
	mutable := extension.Value()
	mutable.Payload.Extensions[0].Name = "tampered"
	if extension.Value().Payload.Extensions[0].Name != "vector" {
		t.Fatal("verified extension receipt exposed mutable storage")
	}
	cleanupDocument := signRuntimeCleanupReceipt(t, signer, validRuntimeCleanupReceipt())
	if !bytes.Contains(cleanupDocument, []byte(`"password_state":null`)) {
		t.Fatal("cleanup receipt did not encode password state as JSON null")
	}
	if _, err := VerifyCredentialCleanupReceipt(cleanupDocument, root, runtimeTestNow); err != nil {
		t.Fatalf("verify cleanup receipt: %v", err)
	}
}

func TestVerifyRuntimeReceiptsRejectStrictShapeTamperAndFutureTime(t *testing.T) {
	root, signer := runtimeTestRootAndSigner(t)
	bootstrap := validRuntimeBootstrapReceipt()
	document := signRuntimeBootstrapReceipt(t, signer, bootstrap)
	unknown := bytes.Replace(
		document, []byte(`{"contract_version":`),
		[]byte(`{"unknown":true,"contract_version":`), 1,
	)
	if _, err := VerifyBootstrapReceipt(unknown, root, runtimeTestNow); err == nil {
		t.Fatal("bootstrap receipt with unknown member accepted")
	}
	bootstrap.Payload.EnvironmentID = "99999999-9999-4999-8999-999999999999"
	if _, err := VerifyBootstrapReceipt(mustRuntimeJSON(t, bootstrap), root, runtimeTestNow); err == nil {
		t.Fatal("tampered bootstrap environment accepted")
	}
	cleanupDocument := signRuntimeCleanupReceipt(t, signer, validRuntimeCleanupReceipt())
	stringNull := bytes.Replace(
		cleanupDocument, []byte(`"password_state":null`),
		[]byte(`"password_state":"null"`), 1,
	)
	if _, err := VerifyCredentialCleanupReceipt(stringNull, root, runtimeTestNow); err == nil {
		t.Fatal("cleanup receipt string null accepted")
	}
	future := validRuntimeExtensionReceipt()
	future.Payload.IssuedAt = "2026-07-13T00:11:00Z"
	if _, err := VerifyExtensionReceipt(
		signRuntimeExtensionReceipt(t, signer, future), root, runtimeTestNow,
	); err == nil {
		t.Fatal("future extension receipt accepted")
	}
}

func signRuntimeBootstrapReceipt(t *testing.T, signer Signer, receipt runtimev1.BootstrapReceipt) []byte {
	t.Helper()
	envelope := bootstrapReceiptEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	}
	receipt.Signature = signRuntimeEnvelope(t, signer, envelope, receipt.Payload.IssuedAt)
	return mustRuntimeJSON(t, receipt)
}

func signRuntimeExtensionReceipt(t *testing.T, signer Signer, receipt runtimev1.ExtensionReceipt) []byte {
	t.Helper()
	envelope := extensionReceiptEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	}
	receipt.Signature = signRuntimeEnvelope(t, signer, envelope, receipt.Payload.IssuedAt)
	return mustRuntimeJSON(t, receipt)
}

func signRuntimeCleanupReceipt(
	t *testing.T, signer Signer, receipt runtimev1.CredentialCleanupReceipt,
) []byte {
	t.Helper()
	envelope := cleanupReceiptEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	}
	receipt.Signature = signRuntimeEnvelope(t, signer, envelope, receipt.Payload.IssuedAt)
	return mustRuntimeJSON(t, receipt)
}

func signRuntimeEnvelope(
	t *testing.T, signer Signer, envelope any, issuedAtValue string,
) runtimev1.AttestationSignature {
	t.Helper()
	issuedAt, err := time.Parse(time.RFC3339Nano, issuedAtValue)
	if err != nil {
		t.Fatalf("parse receipt issue time: %v", err)
	}
	signature, err := signer.SignAt(mustRuntimeJSON(t, envelope), issuedAt)
	if err != nil {
		t.Fatalf("sign receipt envelope: %v", err)
	}
	return runtimeContractSignature(signature)
}
