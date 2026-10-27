package cutover

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type FinalizationEvidence struct {
	AdoptionLease    attestation.VerifiedCutoverCredentialLease
	SessionSeal      attestation.VerifiedCutoverSessionSeal
	Cleanup          attestation.VerifiedCutoverCredentialCleanup
	Authority        attestation.VerifiedAuthorityAdoptionReceipt
	FinalizerLease   attestation.VerifiedCutoverFinalizerLease
	TerminalEvidence attestation.VerifiedCutoverTerminalEvidence
}

func LoadFinalizationEvidence(
	artifacts config.CutoverFinalizationArtifacts,
	now time.Time,
) (FinalizationEvidence, error) {
	leaseDocument, err := readArtifact(artifacts.LeaseManifestPath().Value(), maximumCutoverAuthorityBytes)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	leaseSignature, err := readArtifact(artifacts.LeaseSignaturePath().Value(), maximumCutoverReceiptBytes)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	rootDocument, err := readArtifact(artifacts.TrustRootPath().Value(), maximumCutoverReceiptBytes)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	root, err := attestation.ParseTrustRoot(rootDocument)
	if err != nil {
		return FinalizationEvidence{}, fmt.Errorf("parse cutover trust root: %w", err)
	}
	lease, err := attestation.VerifyCutoverCredentialLease(leaseDocument, leaseSignature, root, now)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	authorityEvidence, err := loadAuthorityEvidence(artifacts, lease, root, now)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	sealDocument, err := readArtifact(artifacts.SessionSealReceiptPath().Value(), maximumCutoverReceiptBytes)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	seal, err := attestation.VerifyCutoverSessionSeal(sealDocument, lease, root, now)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	cleanupDocument, err := readArtifact(artifacts.CredentialCleanupPath().Value(), maximumCutoverReceiptBytes)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	cleanup, err := attestation.VerifyCutoverCredentialCleanup(cleanupDocument, lease, seal, root, now)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	finalizerDocument, err := readArtifact(artifacts.FinalizerLeaseManifestPath().Value(), maximumCutoverReceiptBytes)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	finalizerSignature, err := readArtifact(artifacts.FinalizerLeaseSignaturePath().Value(), maximumCutoverReceiptBytes)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	finalizer, err := attestation.VerifyCutoverFinalizerLease(
		finalizerDocument, finalizerSignature, lease, cleanup,
		authorityEvidence.receipt, root, now,
	)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	terminalDocument, err := readArtifact(
		artifacts.TerminalEvidencePath().Value(), maximumCutoverAuthorityBytes,
	)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	terminal, err := attestation.VerifyCutoverTerminalEvidence(
		terminalDocument, lease, finalizer,
	)
	if err != nil {
		return FinalizationEvidence{}, err
	}
	return FinalizationEvidence{
		AdoptionLease: lease, SessionSeal: seal, Cleanup: cleanup,
		Authority: authorityEvidence.receipt, FinalizerLease: finalizer,
		TerminalEvidence: terminal,
	}, nil
}
