package main

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func buildCleanupReceipt(
	now time.Time, ids releaseIDs, live liveDatabaseIdentity, catalog releaseCatalog,
	authority, release signedArtifactPair, signer signingAuthority,
) (artifact, error) {
	issuedAt := now.Add(-15 * time.Second)
	principals := []runtimev1.CleanedPrincipal{
		{Name: "harpy_cluster_admin", LeaseID: ids.clusterLease,
			CleanedAt: issuedAt.Format(time.RFC3339)},
		{Name: "harpy_extension_admin", LeaseID: ids.extensionLease,
			CleanedAt: issuedAt.Format(time.RFC3339)},
	}
	payload := runtimev1.CredentialCleanupReceiptPayload{
		RunID: ids.cleanupRun, ReleaseID: ids.release,
		IssuedAt:                         issuedAt.Format(time.RFC3339),
		ReleaseManifestSHA256:            release.document.digest,
		BootstrapAuthorityManifestSHA256: authority.document.digest,
		CleanupOperationDigest:           catalog.cleanupOperation.digest,
		DatabaseName:                     live.name, DatabaseOrigin: "fresh",
		DatabaseSystemIdentifier: live.systemID, DatabaseOID: live.databaseOID,
		DatabaseOwner: live.owner, DatabaseEncoding: live.encoding,
		DatabaseLocaleProvider: live.provider, DatabaseCollation: live.collation,
		DatabaseCType: live.ctype, DataChecksums: live.checksums,
		Principals: principals, Result: "succeeded",
	}
	signature, err := receiptSignature(
		signer, "credential-cleanup-receipt.v1", "credential_cleanup", payload, issuedAt,
	)
	if err != nil {
		return artifact{}, err
	}
	receipt := runtimev1.CredentialCleanupReceipt{
		ContractVersion: "credential-cleanup-receipt.v1", Kind: "credential_cleanup",
		Payload: payload, Signature: signature,
	}
	if err := runtimev1.ValidateCredentialCleanupReceipt(receipt); err != nil {
		return artifact{}, fmt.Errorf("validate cleanup receipt: %w", err)
	}
	return newArtifact("cleanup-receipt.json", receipt)
}
