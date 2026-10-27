//go:build integration

package migrator

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func writeBuiltEntrypointCleanupReceipt(
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
	issuedAt := now.Add(-15 * time.Second)
	database := builtEntrypointReceiptDatabase(live)
	payload := runtimev1.CredentialCleanupReceiptPayload{
		RunID:     "77777777-7777-4777-8777-777777777773",
		ReleaseID: builtEntrypointReleaseID, IssuedAt: issuedAt.Format(time.RFC3339),
		ReleaseManifestSHA256:            release.digest,
		BootstrapAuthorityManifestSHA256: authority.digest,
		CleanupOperationDigest:           catalog.cleanupOperation.digest,
		DatabaseName:                     database.DatabaseName,
		DatabaseSystemIdentifier:         database.DatabaseSystemIdentifier,
		DatabaseOID:                      database.DatabaseOID, DatabaseOrigin: database.DatabaseOrigin,
		DatabaseOwner: database.DatabaseOwner, DatabaseEncoding: database.DatabaseEncoding,
		DatabaseLocaleProvider: database.DatabaseLocaleProvider,
		DatabaseCollation:      database.DatabaseCollation, DatabaseCType: database.DatabaseCType,
		DataChecksums: database.DataChecksums,
		Principals: []runtimev1.CleanedPrincipal{
			{Name: "harpy_cluster_admin", LeaseID: builtEntrypointClusterLease,
				CleanedAt: issuedAt.Format(time.RFC3339)},
			{Name: "harpy_extension_admin", LeaseID: builtEntrypointExtensionLease,
				CleanedAt: issuedAt.Format(time.RFC3339)},
		},
		Result: "succeeded",
	}
	receipt := runtimev1.CredentialCleanupReceipt{
		ContractVersion: "credential-cleanup-receipt.v1", Kind: "credential_cleanup",
		Payload: payload,
		Signature: builtEntrypointReceiptSignature(
			t, signer.signer, "credential-cleanup-receipt.v1",
			"credential_cleanup", payload, issuedAt,
		),
	}
	if err := runtimev1.ValidateCredentialCleanupReceipt(receipt); err != nil {
		t.Fatalf("validate built entrypoint cleanup receipt: %v", err)
	}
	return writeBuiltEntrypointArtifact(t, directory, "cleanup-receipt.json", receipt)
}
