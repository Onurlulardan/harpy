package main

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func buildExtensionReceipt(
	now time.Time, ids releaseIDs, live liveDatabaseIdentity, catalog releaseCatalog,
	authority, release signedArtifactPair, signer signingAuthority,
) (artifact, error) {
	issuedAt := now.Add(-30 * time.Second)
	records := make([]runtimev1.ExtensionRecord, 0, len(extensionManifest().Extensions))
	for _, extension := range extensionManifest().Extensions {
		records = append(records, runtimev1.ExtensionRecord{
			Name: extension.Name, Version: extension.Version, Schema: extension.Schema,
			Owner: extension.Owner, ControlFileSHA256: extension.ControlFileSHA256,
		})
	}
	payload := runtimev1.ExtensionReceiptPayload{
		RunID: ids.extensionRun, ReleaseID: ids.release,
		IssuedAt:                         issuedAt.Format(time.RFC3339),
		ReleaseManifestSHA256:            release.document.digest,
		BootstrapAuthorityManifestSHA256: authority.document.digest,
		CredentialLeaseID:                ids.extensionLease,
		PostgresImageDigest:              runtimev1.PostgresImageDigest,
		OwnershipManifestSHA256:          catalog.ownership.digest,
		DatabaseName:                     live.name, DatabaseSystemIdentifier: live.systemID,
		DatabaseOID: live.databaseOID, ServerVersionNumber: live.serverVersion,
		DatabaseOrigin: "fresh", DatabaseOwner: live.owner,
		DatabaseEncoding: live.encoding, DatabaseLocaleProvider: live.provider,
		DatabaseCollation: live.collation, DatabaseCType: live.ctype,
		DataChecksums: live.checksums,
		SessionUser:   "harpy_extension_admin", CurrentUser: "harpy_extension_admin",
		OperationDigest: catalog.extensionOperation.digest, Result: "succeeded",
		ExtensionManifestSHA256: catalog.extension.digest, Extensions: records,
	}
	signature, err := receiptSignature(
		signer, "postgres-extension-receipt.v1", "extension_admin", payload, issuedAt,
	)
	if err != nil {
		return artifact{}, err
	}
	receipt := runtimev1.ExtensionReceipt{
		ContractVersion: "postgres-extension-receipt.v1", Kind: "extension_admin",
		Payload: payload, Signature: signature,
	}
	if err := runtimev1.ValidateExtensionReceipt(receipt); err != nil {
		return artifact{}, fmt.Errorf("validate extension receipt: %w", err)
	}
	return newArtifact("extension-receipt.json", receipt)
}
