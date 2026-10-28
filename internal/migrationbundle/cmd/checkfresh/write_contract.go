package main

import (
	"fmt"
	"os"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func writeFreshArtifacts(root, digest string, manifest []byte) error {
	value, err := loadContract(root)
	if err != nil {
		return err
	}
	value.MigrationManifestSHA256 = digest
	canonical, err := attestation.MarshalCanonical(value)
	if err != nil {
		return err
	}
	if _, err = attestation.VerifyBinarySchemaContract(canonical); err != nil {
		return err
	}
	if err = os.WriteFile(contractPath(root), canonical, 0o644); err != nil {
		return err
	}
	return os.WriteFile(manifestPath(root), manifest, 0o644)
}

func contractDigest(root string) (string, error) {
	value, err := loadContract(root)
	if err != nil {
		return "", err
	}
	return value.MigrationManifestSHA256, nil
}

func loadContract(root string) (runtimev1.BinarySchemaContract, error) {
	document, err := os.ReadFile(contractPath(root))
	if err != nil {
		return runtimev1.BinarySchemaContract{}, err
	}
	verified, err := attestation.VerifyBinarySchemaContract(document)
	if err != nil {
		return runtimev1.BinarySchemaContract{}, err
	}
	value := verified.Value()
	if value.ContractVersion != runtimev1.BinarySchemaContractVersion {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf("binary schema contract version differs")
	}
	return value, nil
}
