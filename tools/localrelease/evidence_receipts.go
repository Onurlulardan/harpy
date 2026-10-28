package main

import (
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func buildAllReceipts(
	now time.Time, ids releaseIDs, live liveDatabaseIdentity, catalog releaseCatalog,
	authority, cutoverOperation signedArtifactPair,
	operation cutoverv1.AuthorityAdoptionOperation,
	release signedArtifactPair, signer signingAuthority,
) (generatedEvidence, error) {
	bootstrap, err := buildBootstrapReceipt(
		now, ids, live, catalog, authority, release, signer,
	)
	if err != nil {
		return generatedEvidence{}, err
	}
	extension, err := buildExtensionReceipt(
		now, ids, live, catalog, authority, release, signer,
	)
	if err != nil {
		return generatedEvidence{}, err
	}
	cleanup, err := buildCleanupReceipt(
		now, ids, live, catalog, authority, release, signer,
	)
	if err != nil {
		return generatedEvidence{}, err
	}
	cutoverReceipt, err := buildCutoverReceipt(
		now, ids, operation, cutoverOperation, authority, release, signer,
	)
	if err != nil {
		return generatedEvidence{}, err
	}
	return generatedEvidence{
		authority: authority, cutoverOperation: cutoverOperation, release: release,
		cutoverReceipt: cutoverReceipt, bootstrapReceipt: bootstrap,
		extensionReceipt: extension, cleanupReceipt: cleanup,
	}, nil
}
