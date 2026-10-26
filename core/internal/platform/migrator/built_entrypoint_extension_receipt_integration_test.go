//go:build integration

package migrator

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func writeBuiltEntrypointExtensionReceipt(
	t *testing.T,
	directory string,
	now time.Time,
	live builtEntrypointLiveIdentity,
	catalog builtEntrypointArtifactCatalog,
	authority builtEntrypointArtifact,
	release builtEntrypointArtifact,
	signer builtEntrypointSigner,
) builtEntrypointArtifact {
	t.Helper()
	issuedAt := now.Add(-30 * time.Second)
	database := builtEntrypointReceiptDatabase(live)
	payload := runtimev1.ExtensionReceiptPayload{
		RunID:     "77777777-7777-4777-8777-777777777772",
		ReleaseID: builtEntrypointReleaseID, IssuedAt: issuedAt.Format(time.RFC3339),
		ReleaseManifestSHA256:            release.digest,
		BootstrapAuthorityManifestSHA256: authority.digest,
		CredentialLeaseID:                builtEntrypointExtensionLease,
		PostgresImageDigest:              runtimev1.PostgresImageDigest,
		OwnershipManifestSHA256:          catalog.ownership.digest,
		DatabaseName:                     database.DatabaseName,
		DatabaseSystemIdentifier:         database.DatabaseSystemIdentifier,
		DatabaseOID:                      database.DatabaseOID, ServerVersionNumber: runtimev1.PostgresServerVersionNumber,
		DatabaseOrigin: database.DatabaseOrigin, DatabaseOwner: database.DatabaseOwner,
		DatabaseEncoding:       database.DatabaseEncoding,
		DatabaseLocaleProvider: database.DatabaseLocaleProvider,
		DatabaseCollation:      database.DatabaseCollation, DatabaseCType: database.DatabaseCType,
		DataChecksums: database.DataChecksums,
		SessionUser:   "harpy_extension_admin", CurrentUser: "harpy_extension_admin",
		OperationDigest: catalog.extensionOperation.digest, Result: "succeeded",
		ExtensionManifestSHA256: catalog.extension.digest,
		Extensions: []runtimev1.ExtensionRecord{
			{Name: "pg_trgm", Version: "1.6", Schema: "platform_ext",
				Owner:             "harpy_extension_admin",
				ControlFileSHA256: integrationDigest("pg-trgm-control")},
			{Name: "vector", Version: "0.8.3", Schema: "platform_ext",
				Owner:             "harpy_extension_admin",
				ControlFileSHA256: integrationDigest("vector-control")},
		},
	}
	receipt := runtimev1.ExtensionReceipt{
		ContractVersion: "postgres-extension-receipt.v1", Kind: "extension_admin",
		Payload: payload,
		Signature: builtEntrypointReceiptSignature(
			t, signer.signer, "postgres-extension-receipt.v1",
			"extension_admin", payload, issuedAt,
		),
	}
	if err := runtimev1.ValidateExtensionReceipt(receipt); err != nil {
		t.Fatalf("validate built entrypoint extension receipt: %v", err)
	}
	return writeBuiltEntrypointArtifact(t, directory, "extension-receipt.json", receipt)
}
