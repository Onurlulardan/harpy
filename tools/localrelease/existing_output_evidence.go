package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type existingEvidenceIdentity struct {
	environmentID, releaseID, authorityDigest string
	systemID                                  string
	databaseOID                               uint32
}

func loadExistingEvidence(
	directory string,
	now time.Time,
	catalog releaseCatalog,
	packageAuthority catalogPackageAuthority,
) (existingEvidenceIdentity, error) {
	if err := verifyOwnedDirectory(directory, filepath.Dir(directory)); err != nil {
		return existingEvidenceIdentity{}, err
	}
	if err := verifyExistingOutputInventory(directory); err != nil {
		return existingEvidenceIdentity{}, err
	}
	root, err := loadCanonicalOutputArtifact(directory, "trust-root.json")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	loadPair := func(document, signature string) (signedArtifactPair, error) {
		left, err := loadCanonicalOutputArtifact(directory, document)
		if err != nil {
			return signedArtifactPair{}, err
		}
		right, err := loadCanonicalOutputArtifact(directory, signature)
		return signedArtifactPair{document: left, signature: right}, err
	}
	authority, err := loadPair("bootstrap-authority.json", "bootstrap-authority.sig")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	operation, err := loadPair("cutover-authority-operation.json", "cutover-authority-operation.sig")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	release, err := loadPair("release.json", "release.sig")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	cutoverReceipt, err := loadPair("cutover-authority-receipt.json", "cutover-authority-receipt.sig")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	bootstrap, err := loadCanonicalOutputArtifact(directory, "bootstrap-receipt.json")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	extension, err := loadCanonicalOutputArtifact(directory, "extension-receipt.json")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	cleanup, err := loadCanonicalOutputArtifact(directory, "cleanup-receipt.json")
	if err != nil {
		return existingEvidenceIdentity{}, err
	}
	evidence := generatedEvidence{
		authority: authority, cutoverOperation: operation, release: release,
		cutoverReceipt: cutoverReceipt, bootstrapReceipt: bootstrap,
		extensionReceipt: extension, cleanupReceipt: cleanup,
	}
	if err := verifyGeneratedEvidence(
		now, catalog, signingAuthority{trustRoot: root}, packageAuthority, evidence,
	); err != nil {
		return existingEvidenceIdentity{}, fmt.Errorf("verify existing release evidence: %w", err)
	}
	return decodeExistingEvidenceIdentity(authority, bootstrap)
}

func loadCanonicalOutputArtifact(directory, name string) (artifact, error) {
	document, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		return artifact{}, fmt.Errorf("read %s: %w", name, err)
	}
	canonical, err := attestation.Canonicalize(document)
	if err != nil || !bytes.Equal(document, canonical) {
		return artifact{}, fmt.Errorf("existing %s is not canonical: %w", name, err)
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		return artifact{}, err
	}
	return artifact{name: name, digest: digest, document: document}, nil
}

func decodeExistingEvidenceIdentity(
	authority signedArtifactPair, bootstrap artifact,
) (existingEvidenceIdentity, error) {
	var manifest runtimev1.BootstrapAuthorityManifest
	if err := attestation.DecodeStrict(authority.document.document, &manifest); err != nil {
		return existingEvidenceIdentity{}, err
	}
	var receipt runtimev1.BootstrapReceipt
	if err := attestation.DecodeStrict(bootstrap.document, &receipt); err != nil {
		return existingEvidenceIdentity{}, err
	}
	return existingEvidenceIdentity{
		environmentID: manifest.EnvironmentID, releaseID: manifest.ReleaseID,
		authorityDigest: authority.document.digest,
		systemID:        receipt.Payload.DatabaseSystemIdentifier,
		databaseOID:     receipt.Payload.DatabaseOID,
	}, nil
}

func verifyExistingOutputInventory(directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("list existing release output: %w", err)
	}
	actual := make([]string, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("existing release output contains a non-regular entry")
		}
		actual = append(actual, entry.Name())
	}
	expected := append([]string(nil), existingOutputNames...)
	sort.Strings(actual)
	sort.Strings(expected)
	if len(actual) != len(expected) {
		return fmt.Errorf("existing release output inventory differs")
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return fmt.Errorf("existing release output inventory differs")
		}
	}
	return nil
}
