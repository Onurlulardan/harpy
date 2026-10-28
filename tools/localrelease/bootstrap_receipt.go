package main

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func buildBootstrapReceipt(
	now time.Time, ids releaseIDs, live liveDatabaseIdentity, catalog releaseCatalog,
	authority, release signedArtifactPair, signer signingAuthority,
) (artifact, error) {
	issuedAt := now.Add(-45 * time.Second)
	payload := runtimev1.BootstrapReceiptPayload{
		RunID: ids.bootstrapRun, ReleaseID: ids.release, EnvironmentID: ids.environment,
		IssuedAt:                         issuedAt.Format(time.RFC3339),
		ReleaseManifestSHA256:            release.document.digest,
		BootstrapAuthorityManifestSHA256: authority.document.digest,
		CredentialLeaseID:                ids.clusterLease,
		PostgresImageDigest:              runtimev1.PostgresImageDigest,
		OwnershipManifestSHA256:          catalog.ownership.digest,
		DatabaseName:                     live.name, DatabaseSystemIdentifier: live.systemID,
		DatabaseOID: live.databaseOID, ServerVersionNumber: live.serverVersion,
		DatabaseOrigin: "fresh", DatabaseOwner: live.owner,
		DatabaseEncoding: live.encoding, DatabaseLocaleProvider: live.provider,
		DatabaseCollation: live.collation, DatabaseCType: live.ctype,
		DataChecksums: live.checksums,
		SessionUser:   "harpy_cluster_admin", CurrentUser: "harpy_cluster_admin",
		OperationDigest: catalog.bootstrapOperation.digest, Result: "succeeded",
		RolesManifestSHA256:      catalog.roles.digest,
		PrivilegesManifestSHA256: catalog.privileges.digest,
	}
	signature, err := receiptSignature(
		signer, "postgres-bootstrap-receipt.v1", "cluster_bootstrap", payload, issuedAt,
	)
	if err != nil {
		return artifact{}, err
	}
	receipt := runtimev1.BootstrapReceipt{
		ContractVersion: "postgres-bootstrap-receipt.v1", Kind: "cluster_bootstrap",
		Payload: payload, Signature: signature,
	}
	if err := runtimev1.ValidateBootstrapReceipt(receipt); err != nil {
		return artifact{}, fmt.Errorf("validate bootstrap receipt: %w", err)
	}
	return newArtifact("bootstrap-receipt.json", receipt)
}
