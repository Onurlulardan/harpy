package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

type generatedEvidence struct {
	authority, cutoverOperation, release, cutoverReceipt signedArtifactPair
	bootstrapReceipt, extensionReceipt, cleanupReceipt   artifact
}

func verifyGeneratedEvidence(
	now time.Time,
	catalog releaseCatalog,
	signer signingAuthority,
	packageAuthority catalogPackageAuthority,
	evidence generatedEvidence,
) error {
	root, err := attestation.ParseTrustRoot(signer.trustRoot.document)
	if err != nil {
		return fmt.Errorf("verify trust root: %w", err)
	}
	release, err := attestation.VerifyRelease(
		evidence.release.document.document, evidence.release.signature.document,
		root, now, databaseName,
	)
	if err != nil {
		return fmt.Errorf("verify release: %w", err)
	}
	if _, err := attestation.VerifyCatalogPackageSet(
		catalog.catalogPackageSet.document,
		release,
		packageAuthority,
		root.ActivePublicKeys(release.IssuedAt()),
	); err != nil {
		return fmt.Errorf("verify Catalog package set: %w", err)
	}
	authority, err := attestation.VerifyBootstrapAuthority(
		evidence.authority.document.document, evidence.authority.signature.document, root, now,
	)
	if err != nil {
		return fmt.Errorf("verify bootstrap authority: %w", err)
	}
	artifacts, err := verifiedPostgresArtifacts(catalog)
	if err != nil {
		return err
	}
	bound, err := attestation.BindPostgresArtifactSet(release, authority, artifacts)
	if err != nil || !bound.IsVerified() {
		return fmt.Errorf("bind PostgreSQL artifact set: %w", err)
	}
	binary, err := attestation.BuildBinarySchemaContract(
		release, artifacts.Migration, artifacts.Extension,
	)
	if err != nil || !bytes.Equal(binary.CanonicalJSON(), catalog.binaryContract.document) {
		return fmt.Errorf("binary schema contract binding differs: %w", err)
	}
	bootstrap, err := attestation.VerifyBootstrapReceipt(
		evidence.bootstrapReceipt.document, root, now,
	)
	if err != nil {
		return err
	}
	extension, err := attestation.VerifyExtensionReceipt(
		evidence.extensionReceipt.document, root, now,
	)
	if err != nil {
		return err
	}
	cleanup, err := attestation.VerifyCredentialCleanupReceipt(
		evidence.cleanupReceipt.document, root, now,
	)
	if err != nil {
		return err
	}
	if err := verifyRuntimeReceiptBindings(
		release.Value(), release.ManifestSHA256(), authority.Value(), catalog,
		bootstrap.Value(), extension.Value(), cleanup.Value(),
	); err != nil {
		return err
	}
	operation, err := attestation.VerifyAuthorityAdoptionOperation(
		evidence.cutoverOperation.document.document,
		evidence.cutoverOperation.signature.document, release, authority, root, now,
	)
	if err != nil {
		return fmt.Errorf("verify cutover authority operation: %w", err)
	}
	if _, err := attestation.VerifyAuthorityAdoptionReceipt(
		evidence.cutoverReceipt.document.document, evidence.cutoverReceipt.signature.document,
		operation, release, authority, root, now,
	); err != nil {
		return fmt.Errorf("verify cutover authority receipt: %w", err)
	}
	return nil
}
