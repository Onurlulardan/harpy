package cutover

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type AdoptionEvidence struct {
	Lease attestation.VerifiedCutoverCredentialLease
	Batch attestation.VerifiedCutoverChangeBatch
	Root  attestation.TrustRoot
}

func LoadAdoptionEvidence(
	artifacts config.CutoverAdoptionArtifacts,
	now time.Time,
) (AdoptionEvidence, error) {
	leaseDocument, err := readArtifact(
		artifacts.LeaseManifestPath().Value(), maximumCutoverAuthorityBytes,
	)
	if err != nil {
		return AdoptionEvidence{}, err
	}
	signatureDocument, err := readArtifact(
		artifacts.LeaseSignaturePath().Value(), maximumCutoverReceiptBytes,
	)
	if err != nil {
		return AdoptionEvidence{}, err
	}
	rootDocument, err := readArtifact(
		artifacts.TrustRootPath().Value(), maximumCutoverReceiptBytes,
	)
	if err != nil {
		return AdoptionEvidence{}, err
	}
	root, err := attestation.ParseTrustRoot(rootDocument)
	if err != nil {
		return AdoptionEvidence{}, fmt.Errorf("parse cutover trust root: %w", err)
	}
	lease, err := attestation.VerifyCutoverCredentialLease(
		leaseDocument, signatureDocument, root, now,
	)
	if err != nil {
		return AdoptionEvidence{}, fmt.Errorf("verify cutover lease: %w", err)
	}
	batchDocument, err := readArtifact(artifacts.BatchPath().Value(), maximumCutoverBatchBytes)
	if err != nil {
		return AdoptionEvidence{}, err
	}
	batch, err := attestation.VerifyCutoverChangeBatch(batchDocument, lease)
	if err != nil {
		return AdoptionEvidence{}, fmt.Errorf("verify cutover batch: %w", err)
	}
	return AdoptionEvidence{Lease: lease, Batch: batch, Root: root}, nil
}

func LoadSessionSeal(
	path string,
	evidence AdoptionEvidence,
	now time.Time,
) (attestation.VerifiedCutoverSessionSeal, error) {
	document, err := readArtifact(path, maximumCutoverReceiptBytes)
	if err != nil {
		return attestation.VerifiedCutoverSessionSeal{}, err
	}
	return attestation.VerifyCutoverSessionSeal(document, evidence.Lease, evidence.Root, now)
}
