package main

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

type finalizationDocuments struct {
	finalizerLease []byte
	adoptionLease  []byte
	seal           []byte
	cleanup        []byte
	authority      []byte
	terminal       []byte
}

type verifiedFinalizationPrerequisites struct {
	authority verifiedAuthorityBundle
	adoption  attestation.VerifiedCutoverCredentialLease
	seal      attestation.VerifiedCutoverSessionSeal
	cleanup   attestation.VerifiedCutoverCredentialCleanup
	receipt   attestation.VerifiedAuthorityAdoptionReceipt
	terminal  issuedTerminalEvidence
	documents finalizationDocuments
}

type verifiedFinalizationBundle struct {
	authority verifiedAuthorityBundle
	adoption  attestation.VerifiedCutoverCredentialLease
	seal      attestation.VerifiedCutoverSessionSeal
	cleanup   attestation.VerifiedCutoverCredentialCleanup
	receipt   attestation.VerifiedAuthorityAdoptionReceipt
	finalizer attestation.VerifiedCutoverFinalizerLease
	terminal  attestation.VerifiedCutoverTerminalEvidence
	documents finalizationDocuments
}

func loadFinalizationBundle(
	paths authorizationInvocation, now time.Time,
) (verifiedFinalizationBundle, error) {
	prerequisites, err := loadFinalizationPrerequisites(
		paths.finalizationPrerequisitePaths, now,
	)
	if err != nil {
		return verifiedFinalizationBundle{}, err
	}
	return loadFinalizerArtifacts(paths, now, prerequisites)
}

func loadFinalizerArtifacts(
	paths authorizationInvocation, now time.Time,
	prerequisites verifiedFinalizationPrerequisites,
) (verifiedFinalizationBundle, error) {
	manifest, err := readArtifact(paths.FinalizerLease, maximumSecretBytes)
	if err != nil {
		return verifiedFinalizationBundle{}, err
	}
	signature, err := readArtifact(paths.FinalizerLeaseSignature, maximumSecretBytes)
	if err != nil {
		return verifiedFinalizationBundle{}, err
	}
	finalizer, err := attestation.VerifyCutoverFinalizerLease(
		manifest, signature, prerequisites.adoption, prerequisites.cleanup,
		prerequisites.receipt, prerequisites.authority.root, now,
	)
	if err != nil {
		return verifiedFinalizationBundle{}, fmt.Errorf("finalizer lease verification failed")
	}
	terminal, err := attestation.VerifyCutoverTerminalEvidence(
		prerequisites.terminal.document, prerequisites.adoption, finalizer,
	)
	if err != nil {
		return verifiedFinalizationBundle{}, fmt.Errorf("terminal evidence verification failed")
	}
	if terminal.Value().TerminalState != finalizer.TerminalState() {
		return verifiedFinalizationBundle{}, fmt.Errorf("terminal evidence state differs")
	}
	documents := prerequisites.documents
	documents.finalizerLease = finalizer.CanonicalDocument()
	documents.terminal = terminal.CanonicalDocument()
	return verifiedFinalizationBundle{authority: prerequisites.authority,
		adoption: prerequisites.adoption, seal: prerequisites.seal,
		cleanup: prerequisites.cleanup, receipt: prerequisites.receipt,
		finalizer: finalizer, terminal: terminal, documents: documents}, nil
}
