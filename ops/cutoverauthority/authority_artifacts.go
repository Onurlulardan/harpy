package main

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

type verifiedAuthorityBundle struct {
	root      attestation.TrustRoot
	release   attestation.VerifiedRelease
	authority attestation.VerifiedBootstrapAuthority
	operation attestation.VerifiedAuthorityAdoptionOperation
}

func loadAuthorityBundle(
	paths authorityArtifactPaths, now time.Time,
) (verifiedAuthorityBundle, error) {
	rootDocument, err := readArtifact(paths.TrustRoot, maximumSecretBytes)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	root, err := attestation.ParseTrustRoot(rootDocument)
	if err != nil {
		return verifiedAuthorityBundle{}, fmt.Errorf("trust root verification failed")
	}
	releaseDocument, err := readArtifact(paths.ReleaseManifest, maximumArtifactBytes)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	releaseSignature, err := readArtifact(paths.ReleaseSignature, maximumSecretBytes)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	var releaseHeader attestation.ReleaseManifest
	if err := attestation.DecodeStrict(releaseDocument, &releaseHeader); err != nil {
		return verifiedAuthorityBundle{}, fmt.Errorf("release artifact verification failed")
	}
	release, err := attestation.VerifyRelease(
		releaseDocument, releaseSignature, root, now, releaseHeader.DatabaseName,
	)
	if err != nil {
		return verifiedAuthorityBundle{}, fmt.Errorf("release artifact verification failed")
	}
	authorityDocument, err := readArtifact(paths.BootstrapAuthority, maximumArtifactBytes)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	authoritySignature, err := readArtifact(
		paths.BootstrapAuthoritySignature, maximumSecretBytes,
	)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	authority, err := attestation.VerifyBootstrapAuthority(
		authorityDocument, authoritySignature, root, now,
	)
	if err != nil {
		return verifiedAuthorityBundle{}, fmt.Errorf("bootstrap authority verification failed")
	}
	operationDocument, err := readArtifact(paths.AuthorityOperation, maximumArtifactBytes)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	operationSignature, err := readArtifact(
		paths.AuthorityOperationSignature, maximumSecretBytes,
	)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	operation, err := attestation.VerifyAuthorityAdoptionOperation(
		operationDocument, operationSignature, release, authority, root, now,
	)
	if err != nil {
		return verifiedAuthorityBundle{}, fmt.Errorf("authority operation verification failed")
	}
	return verifiedAuthorityBundle{
		root: root, release: release, authority: authority, operation: operation,
	}, nil
}

func loadAuthoritySigner(
	path string, root attestation.TrustRoot,
) (attestation.Signer, error) {
	document, err := readSecret(path)
	if err != nil {
		return attestation.Signer{}, err
	}
	signer, err := attestation.ParseSigner(document, root)
	if err != nil {
		return attestation.Signer{}, fmt.Errorf("authority signing key verification failed")
	}
	return signer, nil
}
