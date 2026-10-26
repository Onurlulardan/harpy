package runtimev1

import (
	"strings"
	"testing"
)

func TestValidateTypedReceiptsAcceptExactContracts(t *testing.T) {
	digest := strings.Repeat("a", 64)
	bootstrap := typedContractFixture[BootstrapReceipt](t, validBootstrapReceipt(digest))
	if err := ValidateBootstrapReceipt(bootstrap); err != nil {
		t.Fatalf("valid bootstrap receipt rejected: %v", err)
	}
	extension := typedContractFixture[ExtensionReceipt](t, validExtensionReceipt(digest))
	if err := ValidateExtensionReceipt(extension); err != nil {
		t.Fatalf("valid extension receipt rejected: %v", err)
	}
	cleanup := typedContractFixture[CredentialCleanupReceipt](
		t, validCredentialCleanupReceipt(digest),
	)
	if err := ValidateCredentialCleanupReceipt(cleanup); err != nil {
		t.Fatalf("valid cleanup receipt rejected: %v", err)
	}
}

func TestValidateTypedReceiptsRejectSemanticDrift(t *testing.T) {
	digest := strings.Repeat("a", 64)
	t.Run("bootstrap environment", func(t *testing.T) {
		receipt := typedContractFixture[BootstrapReceipt](t, validBootstrapReceipt(digest))
		receipt.Payload.EnvironmentID = ""
		if err := ValidateBootstrapReceipt(receipt); err == nil {
			t.Fatal("bootstrap receipt without environment accepted")
		}
	})
	t.Run("fresh extension authority", func(t *testing.T) {
		receipt := typedContractFixture[ExtensionReceipt](t, validExtensionReceipt(digest))
		receipt.Payload.Extensions[0].Schema = "public"
		if err := ValidateExtensionReceipt(receipt); err == nil {
			t.Fatal("fresh extension outside platform schema accepted")
		}
	})
	t.Run("sorted extensions", func(t *testing.T) {
		receipt := typedContractFixture[ExtensionReceipt](t, validExtensionReceipt(digest))
		receipt.Payload.Extensions = append(receipt.Payload.Extensions, ExtensionRecord{
			Name: "alpha", Version: "1.0", Schema: "platform_ext",
			Owner: "harpy_extension_admin", ControlFileSHA256: digest,
		})
		if err := ValidateExtensionReceipt(receipt); err == nil {
			t.Fatal("unsorted extension evidence accepted")
		}
	})
	t.Run("cleanup JSON null", func(t *testing.T) {
		receipt := typedContractFixture[CredentialCleanupReceipt](
			t, validCredentialCleanupReceipt(digest),
		)
		state := "null"
		receipt.Payload.Principals[0].PasswordState = &state
		if err := ValidateCredentialCleanupReceipt(receipt); err == nil {
			t.Fatal("string null accepted as cleared password")
		}
	})
	t.Run("cleanup lease uniqueness", func(t *testing.T) {
		receipt := typedContractFixture[CredentialCleanupReceipt](
			t, validCredentialCleanupReceipt(digest),
		)
		receipt.Payload.Principals[1].LeaseID = receipt.Payload.Principals[0].LeaseID
		if err := ValidateCredentialCleanupReceipt(receipt); err == nil {
			t.Fatal("duplicate cleanup lease accepted")
		}
	})
}
