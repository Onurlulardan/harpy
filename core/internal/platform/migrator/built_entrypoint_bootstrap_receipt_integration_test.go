//go:build integration

package migrator

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func writeBuiltEntrypointBootstrapReceipt(
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
	issuedAt := now.Add(-45 * time.Second)
	database := builtEntrypointReceiptDatabase(live)
	payload := runtimev1.BootstrapReceiptPayload{
		RunID:     "77777777-7777-4777-8777-777777777771",
		ReleaseID: builtEntrypointReleaseID, EnvironmentID: builtEntrypointEnvironmentID,
		IssuedAt: issuedAt.Format(time.RFC3339), ReleaseManifestSHA256: release.digest,
		BootstrapAuthorityManifestSHA256: authority.digest,
		CredentialLeaseID:                builtEntrypointClusterLease,
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
		SessionUser:   "harpy_cluster_admin", CurrentUser: "harpy_cluster_admin",
		OperationDigest: catalog.bootstrapOperation.digest, Result: "succeeded",
		RolesManifestSHA256:      catalog.roles.digest,
		PrivilegesManifestSHA256: catalog.privileges.digest,
	}
	receipt := runtimev1.BootstrapReceipt{
		ContractVersion: "postgres-bootstrap-receipt.v1", Kind: "cluster_bootstrap",
		Payload: payload,
		Signature: builtEntrypointReceiptSignature(
			t, signer.signer, "postgres-bootstrap-receipt.v1",
			"cluster_bootstrap", payload, issuedAt,
		),
	}
	if err := runtimev1.ValidateBootstrapReceipt(receipt); err != nil {
		t.Fatalf("validate built entrypoint bootstrap receipt: %v", err)
	}
	return writeBuiltEntrypointArtifact(t, directory, "bootstrap-receipt.json", receipt)
}
