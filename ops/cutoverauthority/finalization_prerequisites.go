package main

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func loadFinalizationPrerequisites(
	paths finalizationPrerequisitePaths, now time.Time,
) (verifiedFinalizationPrerequisites, error) {
	authority, err := loadAuthorityBundle(paths.authorityArtifactPaths, now)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, err
	}
	leaseDocument, err := readArtifact(paths.CredentialLease, maximumArtifactBytes)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, err
	}
	leaseSignature, err := readArtifact(paths.CredentialLeaseSignature, maximumSecretBytes)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, err
	}
	adoption, err := attestation.VerifyCutoverCredentialLease(
		leaseDocument, leaseSignature, authority.root, now,
	)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, fmt.Errorf("cutover lease verification failed")
	}
	receiptDocument, receipt, err := loadAuthorityReceipt(paths, authority, now)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, err
	}
	sealDocument, err := readArtifact(paths.SessionSealReceipt, maximumSecretBytes)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, err
	}
	seal, err := attestation.VerifyCutoverSessionSeal(
		sealDocument, adoption, authority.root, now,
	)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, fmt.Errorf("session seal verification failed")
	}
	cleanupDocument, err := readArtifact(paths.CredentialCleanupReceipt, maximumSecretBytes)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, err
	}
	cleanup, err := attestation.VerifyCutoverCredentialCleanup(
		cleanupDocument, adoption, seal, authority.root, now,
	)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, fmt.Errorf("credential cleanup verification failed")
	}
	terminal, err := loadIssuedTerminalEvidence(paths.TerminalEvidence, adoption)
	if err != nil {
		return verifiedFinalizationPrerequisites{}, err
	}
	return verifiedFinalizationPrerequisites{
		authority: authority, adoption: adoption, seal: seal, cleanup: cleanup,
		receipt: receipt, terminal: terminal, documents: finalizationDocuments{
			adoptionLease: adoption.CanonicalDocument(), seal: seal.CanonicalDocument(),
			cleanup: cleanup.CanonicalDocument(), authority: receiptDocument,
			terminal: terminal.document,
		},
	}, nil
}

func loadAuthorityReceipt(
	paths finalizationPrerequisitePaths, authority verifiedAuthorityBundle, now time.Time,
) ([]byte, attestation.VerifiedAuthorityAdoptionReceipt, error) {
	document, err := readArtifact(paths.AuthorityReceipt, maximumArtifactBytes)
	if err != nil {
		return nil, attestation.VerifiedAuthorityAdoptionReceipt{}, err
	}
	signature, err := readArtifact(paths.AuthorityReceiptSignature, maximumSecretBytes)
	if err != nil {
		return nil, attestation.VerifiedAuthorityAdoptionReceipt{}, err
	}
	receipt, err := attestation.VerifyAuthorityAdoptionReceipt(
		document, signature, authority.operation, authority.release,
		authority.authority, authority.root, now,
	)
	if err != nil {
		return nil, attestation.VerifiedAuthorityAdoptionReceipt{},
			fmt.Errorf("authority receipt verification failed")
	}
	return receipt.CanonicalDocument(), receipt, nil
}
