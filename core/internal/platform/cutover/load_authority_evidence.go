package cutover

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type verifiedAuthorityEvidence struct {
	operation attestation.VerifiedAuthorityAdoptionOperation
	receipt   attestation.VerifiedAuthorityAdoptionReceipt
}

func loadAuthorityEvidence(
	artifacts config.CutoverFinalizationArtifacts,
	lease attestation.VerifiedCutoverCredentialLease,
	root attestation.TrustRoot,
	now time.Time,
) (verifiedAuthorityEvidence, error) {
	releaseDocument, err := readArtifact(
		artifacts.ReleaseManifestPath().Value(), maximumCutoverAuthorityBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	releaseSignature, err := readArtifact(
		artifacts.ReleaseSignaturePath().Value(), maximumCutoverReceiptBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	release, err := attestation.VerifyRelease(
		releaseDocument, releaseSignature, root, now, lease.Target().DatabaseName,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	authorityDocument, err := readArtifact(
		artifacts.BootstrapAuthorityPath().Value(), maximumCutoverAuthorityBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	authoritySignature, err := readArtifact(
		artifacts.BootstrapAuthoritySignaturePath().Value(), maximumCutoverReceiptBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	authority, err := attestation.VerifyBootstrapAuthority(
		authorityDocument, authoritySignature, root, now,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	operationDocument, err := readArtifact(
		artifacts.AuthorityOperationPath().Value(), maximumCutoverAuthorityBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	operationSignature, err := readArtifact(
		artifacts.AuthorityOperationSignaturePath().Value(), maximumCutoverReceiptBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	operation, err := attestation.VerifyAuthorityAdoptionOperation(
		operationDocument, operationSignature, release, authority, root, now,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	receiptDocument, err := readArtifact(
		artifacts.AuthorityReceiptPath().Value(), maximumCutoverAuthorityBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	receiptSignature, err := readArtifact(
		artifacts.AuthorityReceiptSignaturePath().Value(), maximumCutoverReceiptBytes,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	receipt, err := attestation.VerifyAuthorityAdoptionReceipt(
		receiptDocument, receiptSignature, operation, release, authority, root, now,
	)
	if err != nil {
		return verifiedAuthorityEvidence{}, err
	}
	return verifiedAuthorityEvidence{operation: operation, receipt: receipt}, nil
}
