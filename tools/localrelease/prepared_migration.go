package main

import (
	"bytes"
	"fmt"
	"os"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func readEmbeddedBuildContract() (runtimev1.BinarySchemaContract, error) {
	document, err := runtimev1.ReadEmbeddedBinarySchemaContract()
	if err != nil {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf("read embedded build contract: %w", err)
	}
	verified, err := attestation.VerifyBinarySchemaContract(document)
	if err != nil {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf("verify embedded build contract: %w", err)
	}
	return verified.Value(), nil
}

func loadPreparedMigration(
	path string, contract runtimev1.BinarySchemaContract,
) (artifact, error) {
	document, err := os.ReadFile(path)
	if err != nil {
		return artifact{}, fmt.Errorf("read prepared migration manifest: %w", err)
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(document)
	if err != nil {
		return artifact{}, fmt.Errorf("verify prepared migration manifest: %w", err)
	}
	canonical := verified.CanonicalJSON()
	if !bytes.Equal(document, canonical) {
		return artifact{}, fmt.Errorf("prepared migration manifest is not canonical")
	}
	if verified.DigestSHA256() != contract.MigrationManifestSHA256 {
		return artifact{}, fmt.Errorf("prepared migration manifest differs from embedded build contract")
	}
	migration := artifact{
		name: "migration-manifest.json", digest: verified.DigestSHA256(), document: canonical,
	}
	expected, err := localSchemaContract(migration, extensionManifestHash)
	if err != nil {
		return artifact{}, err
	}
	if contract != expected {
		return artifact{}, fmt.Errorf("embedded build contract differs from local release contract")
	}
	return migration, nil
}
