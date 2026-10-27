package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func applySchemaContractManifests(manifests *postgresManifests) error {
	root, err := schemaModuleRoot()
	if err != nil {
		return err
	}
	migrationBody, err := schemaMigrationManifest(root)
	if err != nil {
		return err
	}
	extension, err := canonicalDocument(schemaContractExtensionManifest())
	if err != nil {
		return err
	}
	embedded, err := runtimev1.ReadEmbeddedBinarySchemaContract()
	if err != nil {
		return err
	}
	contract, err := attestation.VerifyBinarySchemaContract(embedded)
	if err != nil {
		return err
	}
	expected := contract.Value()
	sum := sha256.Sum256(migrationBody)
	migrationDigest := hex.EncodeToString(sum[:])
	if migrationDigest != expected.MigrationManifestSHA256 ||
		extension.digest != expected.ExtensionManifestSHA256 {
		return fmt.Errorf(
			"schema contract manifest digest differs: migration %s extension %s",
			migrationDigest, extension.digest,
		)
	}
	manifests.migration = releaseDocument{body: migrationBody, digest: migrationDigest}
	manifests.extension = extension
	return nil
}

func schemaContractExtensionManifest() postgresv1.ExtensionManifest {
	return postgresv1.ExtensionManifest{
		ContractVersion: "postgres-extension-manifest.v1", DatabaseOrigin: "fresh",
		PostgresImageDigest:   postgresv1.PostgresImageDigest,
		PostgresEngineVersion: postgresv1.PostgresEngineVersion,
		ControlPathPolicy: postgresv1.ControlPathPolicy{
			Resolver: "pg_config_sharedir", RelativeDirectory: "extension",
		},
		Extensions: []postgresv1.PostgresExtension{
			{Name: "pg_trgm", Version: "1.6", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "pg_trgm.control",
				ControlFileSHA256: "17c2745da2a706e1443ee363ab7424ced871a9a652e86bf55d25d139f43ee27b",
				Capabilities:      []string{"trigram_search"}},
			{Name: "vector", Version: "0.8.3", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "vector.control",
				ControlFileSHA256: "db4ad0045e6622228471acdcb8fc9aeeec0cd6cf9666569d26f723751ea95fcb",
				Capabilities:      []string{"vector_similarity"}},
		},
	}
}

func schemaModuleRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("module root is unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..")), nil
}
